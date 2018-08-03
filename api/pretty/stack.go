package pretty

import (
	"github.com/rancher/norman/api/access"
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/norman/types/definition"
	"github.com/rancher/rio/pkg/pretty"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1/schema"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

func buildStack(request *types.APIContext, resource *types.RawResource, data map[string]interface{}) {
	prettyStack := pretty.Schemas.Schema(&pretty.Version, client.StackType)
	for name, f := range prettyStack.ResourceFields {
		if !definition.IsMapType(f.Type) {
			continue
		}
		targetType := definition.SubType(f.Type)
		addStackResources(request, targetType, name, resource.ID, data)
	}
}

func addStackResources(request *types.APIContext, resourceType, target, stackID string, data map[string]interface{}) {
	var collection []map[string]interface{}
	err := access.List(request, &schema.Version, resourceType, byStackID(stackID), &collection)
	if err != nil || len(collection) == 0 {
		return
	}

	services := map[string]interface{}{}
	for _, data := range collection {
		name := convert.ToString(data["name"])
		delete(data, "name")
		services[name] = data
	}

	data[target] = services
}

func byStackID(id string) *types.QueryOptions {
	return &types.QueryOptions{
		Conditions: []*types.QueryCondition{
			types.EQ(client.ServiceFieldStackID, id),
		},
	}
}
