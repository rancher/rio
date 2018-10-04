package clientaccess

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/pkg/errors"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/cert"
)

var (
	insecureClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
)

type clientToken struct {
	caHash   string
	username string
	password string
}

func AgentAccessInfoToTempKubeConfig(tempDir, server, token string) (string, error) {
	f, err := ioutil.TempFile(tempDir, "tmp-")
	if err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	_, _, err = accessInfoToKubeConfig(f.Name(), server, token, nil)
	if err != nil {
		os.Remove(f.Name())
	}
	return f.Name(), err
}

func AgentAccessInfoToKubeConfig(destFile, server, token string, override *url.URL) ([]byte, *tls.Certificate, error) {
	return accessInfoToKubeConfig(destFile, server, token, override)
}

type Info struct {
	URL      string `json:"url,omitempty"`
	CACerts  []byte `json:"cacerts,omitempty"`
	username string
	password string
	Token    string `json:"token,omitempty"`
}

func (i *Info) WriteKubeConfig(destFile string) error {
	return clientcmd.WriteToFile(*i.KubeConfig(), destFile)
}

func (i *Info) KubeConfig() *clientcmdapi.Config {
	config := clientcmdapi.NewConfig()

	cluster := clientcmdapi.NewCluster()
	cluster.CertificateAuthorityData = i.CACerts
	cluster.Server = i.URL

	authInfo := clientcmdapi.NewAuthInfo()
	if i.password != "" {
		authInfo.Username = i.username
		authInfo.Password = i.password
	} else if i.Token != "" {
		if username, pass, ok := ParseUsernamePassword(i.Token); ok {
			authInfo.Username = username
			authInfo.Password = pass
		} else {
			authInfo.Token = i.Token
		}
	}

	context := clientcmdapi.NewContext()
	context.AuthInfo = "default"
	context.Cluster = "default"

	config.Clusters["default"] = cluster
	config.AuthInfos["default"] = authInfo
	config.Contexts["default"] = context
	config.CurrentContext = "default"

	return config
}

func ParseAndValidateToken(server, token string) (*Info, error) {
	url, err := url.Parse(server)
	if err != nil {
		return nil, errors.Wrapf(err, "Invalid RIO_URL, failed to parse %s", server)
	}

	if url.Scheme != "https" {
		return nil, fmt.Errorf("only https:// URLs are supported, invalid scheme: %s", server)
	}

	for strings.HasSuffix(url.Path, "/") {
		url.Path = url.Path[:len(url.Path)-1]
	}

	parsedToken, err := parseToken(token)
	if err != nil {
		return nil, err
	}

	cacerts, err := getCACerts(*url)
	if err != nil {
		return nil, err
	}

	if ok, hash, newHash := validateCACerts(cacerts, parsedToken.caHash); !ok {
		return nil, fmt.Errorf("RIO_TOKEN does not match the server %s != %s", hash, newHash)
	}

	if err := validateToken(*url, cacerts, parsedToken.username, parsedToken.password); err != nil {
		return nil, err
	}

	return &Info{
		URL:      url.String(),
		CACerts:  cacerts,
		username: parsedToken.username,
		password: parsedToken.password,
		Token:    token,
	}, nil
}

func accessInfoToKubeConfig(destFile, server, token string, overrideURL *url.URL) ([]byte, *tls.Certificate, error) {
	info, err := ParseAndValidateToken(server, token)
	if err != nil {
		return nil, nil, err
	}

	if overrideURL == nil {
		return nil, nil, info.WriteKubeConfig(destFile)
	}

	u, err := url.Parse(info.URL)
	if err != nil {
		return nil, nil, err
	}

	nodeCert, nodeKey, err := getNodeCertKey(*u, info.username, info.password, info.CACerts)
	if err != nil {
		return nil, nil, err
	}

	certs, err := cert.ParseCertsPEM(nodeCert)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to parse node cert")
	}
	if len(certs) < 2 {
		return nil, nil, fmt.Errorf("expected node cert to end with CA cert, length found %d", len(certs))
	}

	overrideInfo := &Info{
		URL:      overrideURL.String(),
		CACerts:  cert.EncodeCertPEM(certs[1]),
		username: info.username,
		password: info.password,
		Token:    token,
	}

	if err := overrideInfo.WriteKubeConfig(destFile); err != nil {
		return nil, nil, errors.Wrapf(err, "Failed to write to kubeconfig %s", destFile)
	}

	tlsCert, err := tls.X509KeyPair(nodeCert, nodeKey)
	return info.CACerts, &tlsCert, err
}

func validateToken(u url.URL, cacerts []byte, username, password string) error {
	u.Path = "/apis"
	_, err := get(u.String(), getHTTPClient(cacerts), username, password)
	if err != nil {
		return errors.Wrap(err, "token is not valid")
	}
	return nil
}

func validateCACerts(cacerts []byte, hash string) (bool, string, string) {
	if len(cacerts) == 0 && hash == "" {
		return true, "", ""
	}

	digest := sha256.Sum256([]byte(cacerts))
	newHash := hex.EncodeToString(digest[:])
	return hash == newHash, hash, newHash
}

func ParseUsernamePassword(token string) (string, string, bool) {
	parsed, err := parseToken(token)
	if err != nil {
		return "", "", false
	}
	return parsed.username, parsed.password, true
}

func parseToken(token string) (clientToken, error) {
	var result clientToken

	if !strings.HasPrefix(token, "R10") {
		return result, fmt.Errorf("RIO_TOKEN is not a valid token format")
	}

	token = token[3:]

	parts := strings.SplitN(token, "::", 2)
	token = parts[0]
	if len(parts) > 1 {
		result.caHash = parts[0]
		token = parts[1]
	}

	parts = strings.SplitN(token, ":", 2)
	if len(parts) != 2 {
		return result, fmt.Errorf("RIO_TOKEN credentials are the wrong format")
	}

	result.username = parts[0]
	result.password = parts[1]

	return result, nil
}

func getHTTPClient(cacerts []byte) *http.Client {
	if len(cacerts) == 0 {
		return http.DefaultClient
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(cacerts)

	return &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
			TLSClientConfig: &tls.Config{
				RootCAs: pool,
			},
		},
	}
}

func getNodeCertKey(u url.URL, username, password string, cacerts []byte) ([]byte, []byte, error) {
	u.Path = "/node.crt"
	url := u.String()

	nodeCert, err := get(url, getHTTPClient(cacerts), username, password)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to download node cert")
	}

	u.Path = "/node.key"
	url = u.String()
	nodeKey, err := get(url, getHTTPClient(cacerts), username, password)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to download node cert")
	}

	return nodeCert, nodeKey, nil
}

func getCACerts(u url.URL) ([]byte, error) {
	u.Path = "/cacerts"
	url := u.String()

	_, err := get(url, http.DefaultClient, "", "")
	if err == nil {
		return nil, nil
	}

	cacerts, err := get(url, insecureClient, "", "")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get CA certs at %s", url)
	}

	_, err = get(url, getHTTPClient(cacerts), "", "")
	if err != nil {
		return nil, errors.Wrapf(err, "server %s is not trusted", url)
	}

	return cacerts, nil
}

func get(u string, client *http.Client, username, password string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	if username != "" {
		req.SetBasicAuth(username, password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", u, resp.Status)
	}

	return ioutil.ReadAll(resp.Body)
}
