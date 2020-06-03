package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-acme/lego/challenge/dns01"

	"github.com/sirupsen/logrus"
)

const (
	AuthorizationHeader = "Authorization"
	ContentTypeHeader   = "Content-Type"
	ContentTypeJSON     = "application/json"
	RdnsSecretName      = "rdns-token"
	txtPathPattern      = "%s/domain/%s/txt"
)

type DNSProvider struct {
	client DNSClient
}

func NewDNSProvider() (*DNSProvider, error) {
	apiEndpoint := os.Getenv("RDNS_API_ENDPOINT")
	token := os.Getenv("RDNS_TOKEN")
	return NewDNSProviderCredential(apiEndpoint, token)
}

func NewDNSProviderCredential(apiEndpoint, token string) (*DNSProvider, error) {
	if apiEndpoint == "" {
		return nil, fmt.Errorf("rdns api endpoint is empty")
	}

	if token == "" {
		return nil, fmt.Errorf("rdns token is missing")
	}

	dnsClient := DNSClient{
		httpClient: http.DefaultClient,
		base:       apiEndpoint,
		token:      token,
	}
	return &DNSProvider{
		client: dnsClient,
	}, nil
}

func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	return d.client.SetTXTRecord(fqdn, value)
}

func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	return d.client.DeleteDNSRecord(domain)
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return 30 * time.Second, 5 * time.Second
}

type DNSClient struct {
	httpClient *http.Client
	base       string
	token      string
}

func (d *DNSClient) SetTXTRecord(domain, text string) error {
	url := fmt.Sprintf(txtPathPattern, d.base, strings.TrimSuffix(domain, "."))
	payload := map[string]string{
		"text": text,
	}
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return err
	}

	// get txt domain first
	method := ""
	resp, err := d.Do(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		// todo: rdns server need to return better status code
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if strings.Contains(string(data), "failed to filter TXT records") {
			method = http.MethodPost
		} else {
			method = http.MethodPut
		}
	}
	resp.Body.Close()

	resp, err = d.Do(method, url, buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		logrus.Infof("expect 200, got %v. Error: %s", resp.StatusCode, string(data))
		return fmt.Errorf("expect 200, got %v. Error: %s", resp.StatusCode, string(data))
	}
	return nil
}

func (d *DNSClient) Do(method, url string, data io.Reader) (*http.Response, error) {
	request, err := http.NewRequest(method, url, data)
	if err != nil {
		return nil, err
	}
	request.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", d.token))
	request.Header.Set(ContentTypeHeader, ContentTypeJSON)
	resp, err := d.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (d *DNSClient) DeleteDNSRecord(domain string) error {
	url := fmt.Sprintf(txtPathPattern, d.base, domain)
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	request.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", d.token))
	request.Header.Set(ContentTypeHeader, ContentTypeJSON)
	_, err = d.httpClient.Do(request)
	if err != nil {
		return err
	}
	return nil
}
