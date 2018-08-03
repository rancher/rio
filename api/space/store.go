package space

import (
	"github.com/rancher/norman/store/transform"
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/norman/types/values"
	"github.com/rancher/rio/api/named"
	"github.com/rancher/rio/pkg/space"
)

type Store struct {
	types.Store
}

func New(store types.Store) *Store {
	return &Store{
		Store: &named.Store{
			Store: &transform.Store{
				Transformer: func(apiContext *types.APIContext, schema *types.Schema, data map[string]interface{}, opt *types.QueryOptions) (map[string]interface{}, error) {
					labels := convert.ToMapInterface(data["labels"])
					if labels[space.SpaceLabel] != "true" {
						return nil, nil
					}
					delete(labels, space.SpaceLabel)
					return data, nil
				},
				Store: store,
			},
		},
	}
}

func (s *Store) Create(apiContext *types.APIContext, schema *types.Schema, data map[string]interface{}) (map[string]interface{}, error) {
	values.PutValue(data, "true", "labels", space.SpaceLabel)
	return s.Store.Create(apiContext, schema, data)
}

func (s *Store) Update(apiContext *types.APIContext, schema *types.Schema, data map[string]interface{}, id string) (map[string]interface{}, error) {
	labels := convert.ToMapInterface(data["labels"])
	labels[space.SpaceLabel] = "true"
	return s.Store.Update(apiContext, schema, data, id)
}
