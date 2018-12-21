package service

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/rancher/norman/pkg/kv"
	"github.com/rancher/norman/store/transform"
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types/client/rio/v1"
)

var SupportedProtocol = map[string]struct{}{
	"http":  {},
	"http2": {},
	"grpc":  {},
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
	space := ""
	parts := strings.SplitN(projectId, "-", 2)
	if len(parts) == 2 {
		space = parts[1]
	}

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
				"url": appendPort("https", name, stackName, space, domain),
			},
		}
	}

	return data, nil
}

func appendPort(protocol, name, stackName, space, domain string) string {
	url := fmt.Sprintf("%s://%s.%s", protocol, HashIfNeed(name, stackName, space), domain)
	if settings.DefaultHTTPSOpenPort.Get() != "443" {
		return fmt.Sprintf("%s:%s", url, settings.DefaultHTTPSOpenPort.Get())
	}
	return url
}

func HashIfNeed(name, stackName, space string) string {
	fullPath := fmt.Sprintf("%s-%s-%s", name, stackName, space)
	if len(fullPath) > 63 {
		digest := sha256.Sum256([]byte(fullPath))
		return fullPath[0:57] + "-" + string(digest[:])[0:5]
	}
	return fullPath
}
