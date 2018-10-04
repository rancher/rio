package clientcfg

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/rancher/norman/clientbase"
	"github.com/rancher/rio/pkg/clientaccess"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	spaceclient "github.com/rancher/rio/types/client/space/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func (c *clusterClientInfo) url(p string) string {
	newURL := c.serverURL
	newURL.Path = path.Join(c.prefix, p)
	return newURL.String()
}

type clusterClientInfo struct {
	cfg        *rest.Config
	prefix     string
	serverURL  url.URL
	httpClient *http.Client
	wsDialer   *websocket.Dialer
}

func (c *clusterClientInfo) spaceClient() (*spaceclient.Client, error) {
	return spaceclient.NewClient(&clientbase.ClientOpts{
		URL:        c.url("/v1beta1-rio/schemas"),
		HTTPClient: c.httpClient,
		WSDialer:   c.wsDialer,
	})
}

func (c *clusterClientInfo) workspaceClient(space string) (*client.Client, error) {
	return client.NewClient(&clientbase.ClientOpts{
		URL:        c.url("/v1beta1-rio/spaces/" + space + "/schemas"),
		HTTPClient: c.httpClient,
		WSDialer:   c.wsDialer,
	})
}

func (c *clusterClientInfo) Domain() (string, error) {
	req, err := http.NewRequest(http.MethodGet, c.url("/domain"), nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	domain, err := ioutil.ReadAll(resp.Body)
	return string(domain), err
}

func newClusterClientInfo(info *clientaccess.Info) (*clusterClientInfo, error) {
	cc := clientcmd.NewDefaultClientConfig(*info.KubeConfig(), &clientcmd.ConfigOverrides{})
	cfg, err := cc.ClientConfig()
	if err != nil {
		return nil, err
	}

	prefix := ""
	if strings.Contains(cfg.Host, "/") {
		u, err := url.Parse(cfg.Host)
		if err == nil {
			prefix = path.Join(u.Path, prefix)
		}
	}

	rt, err := rest.TransportFor(cfg)
	if err != nil {
		return nil, err
	}

	tls, err := rest.TLSConfigFor(cfg)
	if err != nil {
		return nil, err
	}

	prepare, err := createPrepareFunc(cfg)
	if err != nil {
		return nil, err
	}

	wsDialer := &websocket.Dialer{
		TLSClientConfig: tls,
		Proxy:           prepare,
	}

	if len(prefix) > 1 { // ignore prefix=/
		rt = newCallback(rt, func(req *http.Request) {
			req.Header.Set("X-API-URL-Prefix", prefix)
		})
	}

	url, _, err := rest.DefaultServerURL(cfg.Host, "", schema.GroupVersion{}, true)
	if err != nil {
		return nil, err
	}

	return &clusterClientInfo{
		cfg:    cfg,
		prefix: prefix,
		httpClient: &http.Client{
			Transport: rt,
		},
		wsDialer:  wsDialer,
		serverURL: *url,
	}, nil
}

func createPrepareFunc(cfg *rest.Config) (func(req *http.Request) (*url.URL, error), error) {
	return func(req *http.Request) (*url.URL, error) {
		capture := &rtCapture{}
		rt, err := rest.HTTPWrappersForConfig(cfg, capture)
		if err != nil {
			return nil, err
		}
		_, err = rt.RoundTrip(req)
		*req = *capture.req
		return nil, err
	}, nil
}

type rtCapture struct {
	req *http.Request
}

func (r *rtCapture) RoundTrip(req *http.Request) (*http.Response, error) {
	r.req = req
	return nil, nil
}

type prepareRt struct {
	rt          http.RoundTripper
	prepareFunc func(*http.Request)
}

func newCallback(rt http.RoundTripper, prepare func(*http.Request)) *prepareRt {
	return &prepareRt{
		rt:          rt,
		prepareFunc: prepare,
	}
}

func (p *prepareRt) RoundTrip(req *http.Request) (*http.Response, error) {
	p.prepareFunc(req)
	return p.rt.RoundTrip(req)
}
