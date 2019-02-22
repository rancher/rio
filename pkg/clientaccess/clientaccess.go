package clientaccess

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
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

type Info struct {
	URL      string `json:"url,omitempty"`
	CACerts  []byte `json:"cacerts,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token,omitempty"`
}

func (i *Info) WriteKubeConfig(destFile string) error {
	return clientcmd.WriteToFile(*i.KubeConfig(), destFile)
}

func (i *Info) RestConfig() *rest.Config {
	config := rest.Config{}
	config.Host = i.URL
	config.CAData = i.CACerts
	config.BearerToken = i.Token
	config.Username = i.Username
	config.Password = i.Password
	return &config
}

func (i *Info) KubeConfig() *clientcmdapi.Config {
	config := clientcmdapi.NewConfig()

	cluster := clientcmdapi.NewCluster()
	cluster.CertificateAuthorityData = i.CACerts
	cluster.Server = i.URL

	authInfo := clientcmdapi.NewAuthInfo()
	if i.Username != "" {
		authInfo.Username = i.Username
		authInfo.Password = i.Password
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
