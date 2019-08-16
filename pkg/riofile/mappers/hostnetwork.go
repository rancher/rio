package mappers

import (
	"github.com/rancher/mapper"
)

type HostNetwork struct {
}

func (d HostNetwork) FromInternal(data map[string]interface{}) {
	if _, ok := data["hostNetwork"]; ok {
		delete(data, "hostNetwork")
		data["net"] = "host"
	}
}

func (d HostNetwork) ToInternal(data map[string]interface{}) error {
	if net, ok := data["net"]; ok && net == "host" {
		delete(data, "net")
		data["hostNetwork"] = true
	}
	return nil
}

func (d HostNetwork) ModifySchema(schema *mapper.Schema, schemas *mapper.Schemas) error {
	schema.ResourceFields["net"] = mapper.Field{}
	return nil
}
