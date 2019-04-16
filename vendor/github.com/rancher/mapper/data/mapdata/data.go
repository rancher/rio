package mapdata

import (
	"fmt"

	"github.com/rancher/mapper/convert"
	"github.com/rancher/mapper/data"
)

func Map(data map[string]interface{}) data.Object {
	return &mapData{
		Val: data,
	}
}

type mapData struct {
	Val       interface{}
	fieldName string
	missing   bool
	parent    *mapData
}

func (m *mapData) DeleteField(name string) {
	delete(m.Map(), name)
}

func (m *mapData) Delete() {
	if m.parent != nil && m.fieldName != "" {
		m.parent.DeleteField(m.fieldName)
	}
}

func (m *mapData) Field(name string) data.Object {
	v, ok := convert.ToMapInterface(m)[name]
	return &mapData{
		Val:       v,
		fieldName: name,
		missing:   !ok,
		parent:    m,
	}
}

func (m *mapData) SetField(field string, val interface{}) data.Object {
	return m.Field(field).Set(val)
}

func (m *mapData) Set(val interface{}) data.Object {
	if obj, ok := val.(*mapData); ok && obj.missing {
		return &mapData{
			missing: true,
		}
	}

	if obj, ok := val.(data.Object); ok {
		val = obj.Interface()
	}

	if m.fieldName == "" || m.parent == nil {
		return &mapData{
			missing: true,
		}
	}

	if m.parent.missing {
		m.parent.Set(map[string]interface{}{
			m.fieldName: val,
		})
		return m
	}

	p := m.parent.Map()
	if p != nil {
		p[m.fieldName] = val
	}

	return m
}

func (m *mapData) Exists() bool {
	return !m.missing
}

func (m *mapData) String() string {
	return fmt.Sprint(m.Val)
}

func (m *mapData) Map() map[string]interface{} {
	return convert.ToMapInterface(m.Val)
}

func (m *mapData) Interface() interface{} {
	return m.Val
}
