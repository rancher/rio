package approuter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"sort"
	"sync"

	"github.com/rancher/rdns-server/model"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	contentType     = "Content-Type"
	jsonContentType = "application/json"
	secretKey       = "rdns-token"
	cnamePath = "/cname"
)

func jsonBody(payload interface{}) (io.Reader, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(payload)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

type SecretLister interface {
	Get(namespace, name string) (*k8scorev1.Secret, error)
}

type SecretCreator interface {
	Create(*k8scorev1.Secret) (*k8scorev1.Secret, error)
	Update(*k8scorev1.Secret) (*k8scorev1.Secret, error)
}

type Client struct {
	httpClient             *http.Client
	base                   string
	lock                   *sync.RWMutex
	managementSecretLister SecretLister
	secrets                SecretCreator
	clusterName            string
}

func (c *Client) request(method string, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add(contentType, jsonContentType)

	return req, nil
}

func (c *Client) do(req *http.Request) (model.Response, error) {
	var data model.Response
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return data, err
	}
	// when err is nil, resp contains a non-nil resp.Body which must be closed
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return data, errors.Wrap(err, "read response body error")
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return data, errors.Wrapf(err, "decode response error: %s", string(body))
	}
	logrus.Debugf("got response entry: %+v", data)
	if code := resp.StatusCode; code < 200 || code > 300 {
		if data.Message != "" {
			return data, errors.Errorf("got request error: %s", data.Message)
		}
	}

	return data, nil
}

func (c *Client) ApplyDomain(hosts []string, subDomain map[string][]string, cname bool) (bool, string, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	d, err := c.GetDomain(cname)
	if err != nil {
		return false, "", err
	}

	if d == nil {
		logrus.Debugf("fqdn configuration does not exist, need to create a new one")
		fqdn, err := c.CreateDomain(hosts, cname)
		return true, fqdn, err
	}

	sort.Strings(d.Hosts)
	sort.Strings(hosts)
	if !reflect.DeepEqual(d.Hosts, hosts) || !reflect.DeepEqual(d.SubDomain, subDomain) {
		logrus.Debugf("fqdn %s or subdomains %+v has some changes, need to update", d.Fqdn, d.SubDomain)
		fqdn, err := c.UpdateDomain(hosts, subDomain, cname)
		return false, fqdn, err
	}
	logrus.Debugf("fqdn %s has no changes, no need to update", d.Fqdn)
	fqdn, _, _ := c.getSecret()

	return false, fqdn, nil
}

func (c *Client) GetDomain(cname bool) (d *model.Domain, err error) {
	fqdn, token, err := c.getSecret()
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "GetDomain: failed to get stored secret")
	}

	path := ""
	if cname {
		path = cnamePath
	}
	url := buildURL(c.base, "/"+fqdn, path)
	req, err := c.request(http.MethodGet, url, nil)
	if err != nil {
		return d, errors.Wrap(err, "GetDomain: failed to build a request")
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	o, err := c.do(req)
	if err != nil {
		return d, errors.Wrap(err, "GetDomain: failed to execute a request")
	}

	if o.Data.Fqdn == "" {
		return nil, nil
	}

	return &o.Data, nil
}

func (c *Client) CreateDomain(hosts []string, cname bool) (string, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	path := ""
	options := &model.DomainOptions{}
	if cname {
		options.CNAME = hosts[0]
		path = cnamePath
	} else {
		options.Hosts = hosts
	}
	url := buildURL(c.base, "", path)
	body, err := jsonBody(options)
	if err != nil {
		return "", err
	}

	req, err := c.request(http.MethodPost, url, body)
	if err != nil {
		return "", errors.Wrap(err, "CreateDomain: failed to build a request")
	}

	resp, err := c.do(req)
	if err != nil {
		return "", errors.Wrap(err, "CreateDomain: failed to execute a request")
	}

	if err = c.setSecret(&resp); err != nil {
		return "", err
	}

	return resp.Data.Fqdn, err
}

func (c *Client) UpdateDomain(hosts []string, subDomain map[string][]string, cname bool) (string, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	fqdn, token, err := c.getSecret()
	if err != nil {
		return "", errors.Wrap(err, "UpdateDomain: failed to get stored secret")
	}

	path := ""
	options := &model.DomainOptions{
		SubDomain: subDomain,
	}
	if cname {
		options.CNAME = hosts[0]
		path = cnamePath
	} else {
		options.Hosts = hosts
	}
	
	url := buildURL(c.base, "/"+fqdn, path)
	body, err := jsonBody(options)
	if err != nil {
		return "", err
	}

	req, err := c.request(http.MethodPut, url, body)
	if err != nil {
		return "", errors.Wrap(err, "UpdateDomain: failed to build a request")
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	_, err = c.do(req)
	if err != nil {
		return "", errors.Wrap(err, "UpdateDomain: failed to execute a request")
	}

	return fqdn, nil
}

func (c *Client) DeleteDomain() (string, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	fqdn, token, err := c.getSecret()
	if err != nil {
		return "", errors.Wrap(err, "DeleteDomain: failed to get stored secret")
	}

	url := buildURL(c.base, "/"+fqdn, "")
	req, err := c.request(http.MethodDelete, url, nil)
	if err != nil {
		return "", errors.Wrap(err, "DeleteDomain: failed to build a request")
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	_, err = c.do(req)
	if err != nil {
		return "", errors.Wrap(err, "DeleteDomain: failed to execute a request")
	}

	return fqdn, err
}

func (c *Client) RenewDomain() (string, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	fqdn, token, err := c.getSecret()
	if err != nil {
		return "", errors.Wrap(err, "RenewDomain: failed to get stored secret")
	}

	url := buildURL(c.base, "/"+fqdn, "/renew")
	req, err := c.request(http.MethodPut, url, nil)
	if err != nil {
		return "", errors.Wrap(err, "RenewDomain: failed to build a request")
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	_, err = c.do(req)
	if err != nil {
		return "", errors.Wrap(err, "RenewDomain: failed to execute a request")
	}

	return fqdn, err
}

func (c *Client) SetBaseURL(base string) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if base != c.base {
		c.base = base
	}
}

func (c *Client) setSecret(resp *model.Response) error {
	s := &k8scorev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretKey,
			Namespace: c.clusterName,
		},
		Type: k8scorev1.SecretTypeOpaque,
		StringData: map[string]string{
			"token": resp.Token,
			"fqdn":  resp.Data.Fqdn,
		},
	}
	_, err := c.secrets.Create(s)

	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return err
	}

	if err != nil && k8serrors.IsAlreadyExists(err) {
		if _, err := c.secrets.Update(s); err != nil {
			return err
		}
		return nil
	}

	return nil
}

//getSecret return token and fqdn
func (c *Client) getSecret() (string, string, error) {
	sec, err := c.managementSecretLister.Get(c.clusterName, secretKey)
	if err != nil {
		return "", "", err
	}
	return string(sec.Data["fqdn"]), string(sec.Data["token"]), nil
}

func NewClient(secrets SecretCreator, secretLister SecretLister, clusterName string) *Client {
	return &Client{
		httpClient:             http.DefaultClient,
		lock:                   &sync.RWMutex{},
		secrets:                secrets,
		managementSecretLister: secretLister,
		clusterName:            clusterName,
	}
}

//buildUrl return request url
func buildURL(base, fqdn, path string) (url string) {
	return fmt.Sprintf("%s/domain%s%s", base, fqdn, path)
}
