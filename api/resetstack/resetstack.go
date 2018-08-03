package resetstack

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1/schema"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

func New(s types.Store) types.Store {
	return &Store{
		Store: s,
	}
}

type Store struct {
	types.Store
}

func (s *Store) Create(apiContext *types.APIContext, schema *types.Schema, data map[string]interface{}) (map[string]interface{}, error) {
	data, err := s.Store.Create(apiContext, schema, data)
	return s.clearStack(apiContext, data, err)
}

func (s *Store) Update(apiContext *types.APIContext, schema *types.Schema, data map[string]interface{}, id string) (map[string]interface{}, error) {
	data, err := s.Store.Update(apiContext, schema, data, id)
	return s.clearStack(apiContext, data, err)
}

func (s *Store) Delete(apiContext *types.APIContext, schema *types.Schema, id string) (map[string]interface{}, error) {
	data, err := s.Store.Delete(apiContext, schema, id)
	return s.clearStack(apiContext, data, err)
}

func (s *Store) clearStack(apiContext *types.APIContext, data map[string]interface{}, err error) (map[string]interface{}, error) {
	if err != nil {
		return data, err
	}

	stackID, _ := data[client.ServiceFieldStackID].(string)
	if stackID == "" {
		return data, err
	}

	stackSchema := apiContext.Schemas.Schema(&schema.Version, client.StackType)
	_, err = stackSchema.Store.Update(apiContext, stackSchema, map[string]interface{}{
		client.StackFieldTemplate: "",
	}, stackID)

	return data, err
}
