package mapper

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/norman/types/mapper"
	"github.com/sirupsen/logrus"
)

type MaybeStringer interface {
	MaybeString() interface{}
}

type NewObject func() MaybeStringer
type ToObject func(interface{}) (interface{}, error)

type ObjectsToSlice struct {
	Field     string
	NewObject NewObject
	ToObject  ToObject
}

func (p ObjectsToSlice) FromInternal(data map[string]interface{}) {
	if data == nil {
		return
	}

	objs, ok := data[p.Field]
	if !ok {
		return
	}

	var result []interface{}
	for _, obj := range convert.ToMapSlice(objs) {
		target := p.NewObject()
		if err := convert.ToObj(obj, target); err != nil {
			logrus.Errorf("Failed to unmarshal slice to object: %v", err)
			continue
		}

		result = append(result, target.MaybeString())
	}

	if len(result) == 0 {
		delete(data, p.Field)
	} else {
		data[p.Field] = result
	}
}

func (p ObjectsToSlice) ToInternal(data map[string]interface{}) error {
	if data == nil {
		return nil
	}

	d, ok := data[p.Field]
	if !ok {
		return nil
	}

	slc, ok := d.([]interface{})
	if !ok {
		return nil
	}

	var newSlc []interface{}

	for _, obj := range slc {
		n, err := convert.ToNumber(obj)
		if err == nil && n > 0 {
			obj = convert.ToString(n)
		}
		newObj, err := p.ToObject(obj)
		if err != nil {
			return err
		}

		if _, isMap := newObj.(map[string]interface{}); !isMap {
			newObj, err = convert.EncodeToMap(newObj)
		}

		newSlc = append(newSlc, newObj)
	}

	data[p.Field] = newSlc
	return nil
}

func (p ObjectsToSlice) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return mapper.ValidateField(p.Field, schema)
}
