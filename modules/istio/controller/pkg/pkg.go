package pkg

import (
	"net/url"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

func Domains(router *riov1.Router) (result []string, err error) {
	seen := map[string]bool{}
	for _, endpoint := range router.Status.Endpoints {
		u, err := url.Parse(endpoint)
		if err != nil {
			return nil, err
		}

		if seen[u.Host] {
			continue
		}
		seen[u.Host] = true

		result = append(result, u.Host)
	}

	return
}
