package defaults

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
)

type DefaultStatusStore struct {
	types.Store
	Default interface{}
}

func (d *DefaultStatusStore) Create(apiContext *types.APIContext, schema *types.Schema, data map[string]interface{}) (map[string]interface{}, error) {
	var err error

	if data == nil {
		data = map[string]interface{}{}
	}

	data["status"], err = convert.EncodeToMap(d.Default)
	if err != nil {
		return nil, err
	}

	return d.Store.Create(apiContext, schema, data)
}
