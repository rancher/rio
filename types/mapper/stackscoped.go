package mapper

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/norman/types/values"
	"github.com/rancher/rio/pkg/namespace"
)

type StackScoped struct {
}

func (s *StackScoped) FromInternal(data map[string]interface{}) {
	labels := convert.ToMapInterface(values.GetValueN(data, "metadata", "labels"))
	if len(labels) > 0 {
		delete(labels, "apply.cattle.io/generationID")
		delete(labels, "apply.cattle.io/groupID")
	}
}

func (s *StackScoped) ToInternal(data map[string]interface{}) error {
	_, nsOk := values.GetValue(data, "metadata", "namespace")
	stackName, stackOk := values.GetValue(data, "spec", "stackId")
	projectName, projectOk := values.GetValue(data, "spec", "projectId")

	if !nsOk && stackOk && projectOk {
		values.PutValue(data, namespace.StackNamespace(convert.ToString(projectName), convert.ToString(stackName)),
			"metadata", "namespace")
	}
	return nil
}

func (s *StackScoped) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return nil
}
