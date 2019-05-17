package mappers

import (
	"fmt"
	"strconv"
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
	minValue, _ := convert.ToNumber(min)
	maxValue, _ := convert.ToNumber(max)
	if maxOk && max != min {
		data["scale"] = fmt.Sprintf("%v-%v", minValue, maxValue)
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
			min, err := strconv.Atoi(parts[0])
			if err != nil {
				return err
			}
			max, err := strconv.Atoi(parts[1])
			if err != nil {
				return err
			}
			data["minScale"] = min
			data["maxScale"] = max
			if min == max {
				data["scale"] = min
			} else {
				data["scale"] = 0
			}
		}
	}

	return nil
}
