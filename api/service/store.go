package service

import (
	"fmt"

	"github.com/rancher/norman/pkg/kv"
	"github.com/rancher/norman/store/transform"
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

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

	expose := false
	for _, ports := range convert.ToMapSlice(data[client.ServiceFieldPortBindings]) {
		proto, _ := ports[client.PortBindingFieldProtocol].(string)
		if proto != "http" {
			continue
		}
		expose = true
		break
	}

	if expose {
		data["endpoints"] = []interface{}{
			map[string]interface{}{
				"url": fmt.Sprintf("http://%s.%s.%s", name, stackName, domain),
			},
		}
	}

	return data, nil
}
