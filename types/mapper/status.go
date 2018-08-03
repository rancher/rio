package mapper

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/norman/types/values"
)

type Status struct {
}

func (s Status) FromInternal(data map[string]interface{}) {
	if "active" != data["state"] {
		return
	}

	if scaleIsZero(data) {
		data["state"] = "inactive"
	}
}

func scaleIsZero(data map[string]interface{}) bool {
	if data["type"] != "/v1beta1-rio/spaces/schemas/service" {
		return false
	}

	ready := values.GetValueN(data, "scaleStatus", "ready")
	available := values.GetValueN(data, "scaleStatus", "available")
	unavailable := values.GetValueN(data, "scaleStatus", "unavailable")
	updated := values.GetValueN(data, "scaleStatus", "updated")
	scale := values.GetValueN(data, "scale")

	for _, v := range []interface{}{ready, available, unavailable, updated, scale} {
		if v == nil {
			continue
		}
		if n, err := convert.ToNumber(v); err != nil || n != 0 {
			return false
		}
	}

	return true
}

func (s Status) ToInternal(data map[string]interface{}) error {
	return nil
}

func (s Status) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return nil
}
