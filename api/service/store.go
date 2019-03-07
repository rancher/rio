package service

import (
	"fmt"

	"github.com/rancher/norman/pkg/kv"
	"github.com/rancher/norman/store/transform"
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/rio/pkg/namespace"
	"github.com/rancher/rio/pkg/settings"
	client "github.com/rancher/rio/types/client/rio/v1"
)

var SupportedProtocol = map[string]struct{}{
	"http":  {},
	"http2": {},
	"grpc":  {},
	"tcp":   {},
}

func New(store types.Store) types.Store {
	return &transform.Store{
		Store:       store,
		Transformer: addEndpoint,
	}
}

func addEndpoint(apiContext *types.APIContext, schema *types.Schema, data map[string]interface{}, opt *types.QueryOptions) (map[string]interface{}, error) {
	domain := settings.ClusterDomain.Get()
	if domain == "" {
		return data, nil
	}

	name, _ := data["name"].(string)
	if name == "" {
		return data, nil
	}

	stackName, _ := data["stackId"].(string)
	_, stackName = kv.Split(stackName, ":")
	if stackName == "" {
		return data, nil
	}

	projectId, _ := data["projectId"].(string)

	expose := false
	for _, ports := range convert.ToMapSlice(data[client.ServiceFieldPortBindings]) {
		proto, _ := ports[client.PortBindingFieldProtocol].(string)
		if _, ok := SupportedProtocol[proto]; !ok {
			continue
		}
		expose = true
		break
	}

	if expose {
		data["endpoints"] = []interface{}{
			map[string]interface{}{
				"url": appendPort("https", name, stackName, projectId, domain),
			},
		}
	}

	return data, nil
}

func appendPort(protocol, name, stackName, space, domain string) string {
	url := fmt.Sprintf("%s://%s.%s", protocol, namespace.HashIfNeed(name, stackName, space), domain)
	if settings.DefaultHTTPSOpenPort.Get() != "443" {
		return fmt.Sprintf("%s:%s", url, settings.DefaultHTTPSOpenPort.Get())
	}
	return url
}
