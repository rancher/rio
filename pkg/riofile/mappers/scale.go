package mappers

import (
	"fmt"
	"strings"

	"github.com/rancher/mapper/convert"
	"github.com/rancher/mapper/mappers"
)

type Scale struct {
	mappers.DefaultMapper
}

func (d Scale) FromInternal(data map[string]interface{}) {
	min, minOk := data["minScale"]
	max, maxOk := data["maxScale"]
	if !minOk {
		min = 0
	}
	if maxOk {
		data["scale"] = fmt.Sprintf("%d-%d", min, max)
	}
}

func (d Scale) ToInternal(data map[string]interface{}) error {
	v, ok := data["scale"]
	if !ok {
		return nil
	}

	scale := convert.ToString(v)
	if strings.Contains(scale, "-") {
		parts := strings.Split(scale, "-")
		if len(parts) == 2 {
			data["minScale"] = parts[0]
			data["maxScale"] = parts[1]
		}
	}

	return nil
}
