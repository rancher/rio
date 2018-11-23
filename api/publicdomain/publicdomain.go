package publicdomain

import (
	"fmt"

	"github.com/rancher/rio/pkg/settings"

	"github.com/rancher/norman/api/access"
	"github.com/rancher/norman/httperror"
	"github.com/rancher/norman/types"
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
	var (
		datas []map[string]interface{}
	)

	domainName, _ := data["domainName"].(string)
	if err := access.List(apiContext, &schema.Version, schema.ID, &types.QueryOptions{
		Conditions: []*types.QueryCondition{
			types.EQ("domainName", domainName),
		},
	}, &datas); err != nil {
		return nil, err
	}
	if len(datas) > 0 {
		return nil, httperror.NewAPIError(httperror.Conflict, fmt.Sprintf("domainName %s is not unique", domainName))
	}
	data["namespace"] = settings.RioSystemNamespace
	return s.Store.Create(apiContext, schema, data)
}
