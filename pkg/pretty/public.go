package pretty

import (
	"fmt"

	"github.com/rancher/mapper/convert"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/apis/rio.cattle.io/v1/schema"
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
	schema := schema.Schemas.Schema(&schema.Version, "internalStack")
	err := schema.Mapper.ToInternal(data)
	return data, err
}

func ToPrettyStack(data []byte) (*v1.StackFile, error) {
	data, err := NormalizeData(StackType, data)
	if err != nil {
		return nil, err
	}

	data, err = internalizeStack(data)
	if err != nil {
		return nil, err
	}

	stack := &v1.StackFile{}
	err = convert.ToObj(data, stack)
	return stack, err
}
