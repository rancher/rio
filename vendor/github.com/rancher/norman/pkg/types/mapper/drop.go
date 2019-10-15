package mapper

import (
	"fmt"

	"github.com/rancher/norman/pkg/data"
	"github.com/rancher/norman/pkg/types"
)

type Drop struct {
	Field    string
	Optional bool
}

func (d Drop) FromInternal(data data.Object) {
	delete(data, d.Field)
}

func (d Drop) ToInternal(data data.Object) error {
	return nil
}

func (d Drop) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	if _, ok := schema.ResourceFields[d.Field]; !ok {
		if !d.Optional {
			return fmt.Errorf("can not drop missing field %s on %s", d.Field, schema.ID)
		}
	}

	delete(schema.ResourceFields, d.Field)
	return nil
}
