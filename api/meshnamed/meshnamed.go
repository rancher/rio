package meshnamed

import (
	"fmt"

	"github.com/rancher/norman/api/access"
	"github.com/rancher/norman/httperror"
	"github.com/rancher/norman/types"
	"github.com/rancher/rio/types/client/rio/v1"
)

type Store struct {
	types.Store
}

func New(store types.Store) *Store {
	return &Store{
		Store: store,
	}
}

func (s *Store) Create(apiContext *types.APIContext, schema *types.Schema, data map[string]interface{}) (map[string]interface{}, error) {
	stackID, _ := data["stackId"].(string)
	name, _ := data["name"].(string)
	for _, s := range []string{client.RouteSetType, client.ServiceType, client.ExternalServiceType} {
		datas := []map[string]interface{}{}
		err := access.List(apiContext, &schema.Version, s, &types.QueryOptions{
			Conditions: []*types.QueryCondition{
				types.EQ("name", name),
				types.EQ("stackId", stackID),
			},
		}, &datas)
		if err != nil {
			return nil, err
		}
		if len(datas) > 0 {
			return nil, httperror.NewAPIError(httperror.Conflict, fmt.Sprintf("name %s is not unique and already exists in %s", name, s))
		}
	}
	return s.Store.Create(apiContext, schema, data)
}
