package pretty

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/rio/pkg/pretty"
	"github.com/rancher/rio/pkg/yaml"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

func InputFormatter(request *types.APIContext, schema *types.Schema, data map[string]interface{}, create bool) error {
	if create {
		return nil
	}

	if request.Option("edited") != "true" {
		return nil
	}

	data, err := pretty.NormalizeData(pretty.SchemaType(schema.ID), data)
	return err
}

func Format(request *types.APIContext, resource *types.RawResource) {
	if request.Option("export") != "true" && request.Option("edit") != "true" {
		return
	}

	data := map[string]interface{}{}
	if resource.Type == client.StackType {
		buildStack(request, resource, data)
	} else {
		data = resource.Values
	}

	pretty.ToPretty(pretty.SchemaType(resource.Type), data)

	if request.ResponseFormat == "yaml" {
		data = yaml.ConvertYAMLKeys(data)
	}

	*resource = types.RawResource{
		Schema:       resource.Schema,
		DropReadOnly: true,
		Values:       data,
	}
}
