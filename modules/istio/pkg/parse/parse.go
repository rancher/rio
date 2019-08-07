package parse

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/rancher/rio/pkg/constants"
)

func TargetURL(target string) (*url.URL, error) {
	if !strings.Contains(target, "://") {
		target = "http://" + target
	}
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func FormatEndpoint(protocol string, endpoints []string) []string {
	for i, endpoint := range endpoints {
		if protocol == "http" && constants.DefaultHTTPOpenPort != "80" {
			endpoints[i] = fmt.Sprintf("%s:%s", endpoint, constants.DefaultHTTPOpenPort)
		}

		if protocol == "https" && constants.DefaultHTTPSOpenPort != "443" {
			endpoints[i] = fmt.Sprintf("%s:%s", endpoint, constants.DefaultHTTPSOpenPort)
		}
	}
	return endpoints
}
