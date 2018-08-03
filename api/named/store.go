package named

import (
	"fmt"
	"math/rand"
	"time"

	"strings"

	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/rancher/norman/api/access"
	"github.com/rancher/norman/httperror"
	"github.com/rancher/norman/types"
)

func init() {
	// RandomName uses the global random
	rand.Seed(int64(time.Now().Nanosecond()))
}

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
		err   error
	)

	stackID, _ := data["stackId"].(string)
	name, _ := data["name"].(string)
	gen := false
	if name == "" {
		gen = true
	}

	for i := 0; i < 3; i++ {
		if gen {
			name = strings.Replace(namesgenerator.GetRandomName(i), "_", "-", -1)
		}
		err = access.List(apiContext, &schema.Version, schema.ID, &types.QueryOptions{
			Conditions: []*types.QueryCondition{
				types.EQ("name", name),
				types.EQ("stackId", stackID),
			},
		}, &datas)
		if err != nil {
			return nil, err
		}
		if !gen || len(datas) == 0 {
			break
		}
	}

	if len(datas) > 0 {
		return nil, httperror.NewAPIError(httperror.Conflict, fmt.Sprintf("name %s is not unique", name))
	}

	data["name"] = name
	return s.Store.Create(apiContext, schema, data)
}

func (s *Store) Update(apiContext *types.APIContext, schema *types.Schema, data map[string]interface{}, id string) (map[string]interface{}, error) {
	delete(data, "name")
	return s.Store.Update(apiContext, schema, data, id)
}
