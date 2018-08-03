package pretty

import (
	"fmt"

	"github.com/rancher/norman/types/convert"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1/schema"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

var (
	StackType   = SchemaType("stack")
	ServiceType = SchemaType("service")
	VolumeType  = SchemaType("volume")
)

type SchemaType string

func NormalizeData(schemaType SchemaType, data map[string]interface{}) (map[string]interface{}, error) {
	schema := Schemas.Schema(&Version, string(schemaType))
	if schema == nil {
		return nil, fmt.Errorf("failed to find %s", schemaType)
	}
	err := schema.Mapper.ToInternal(data)
	return data, err
}

func ToNormalizedStack(data map[string]interface{}) (*Stack, error) {
	data, err := NormalizeData(StackType, data)
	if err != nil {
		return nil, err
	}

	stack := &Stack{}
	err = convert.ToObj(data, stack)
	return stack, err
}

func ToPretty(schemaType SchemaType, data map[string]interface{}) (map[string]interface{}, error) {
	schema := Schemas.Schema(&Version, string(schemaType))
	if schema == nil {
		return nil, fmt.Errorf("failed to find %s", schemaType)
	}
	schema.Mapper.FromInternal(data)
	return data, nil
}

func internalizeStack(data map[string]interface{}) (map[string]interface{}, error) {
	schema := schema.Schemas.Schema(&schema.Version, client.InternalStackType)
	err := schema.Mapper.ToInternal(data)
	return data, err
}

func ToInternalStack(data map[string]interface{}) (*v1beta1.InternalStack, error) {
	data, err := NormalizeData(StackType, data)
	if err != nil {
		return nil, err
	}

	data, err = internalizeStack(data)
	if err != nil {
		return nil, err
	}

	stack := &v1beta1.InternalStack{}
	err = convert.ToObj(data, stack)
	return stack, err
}
