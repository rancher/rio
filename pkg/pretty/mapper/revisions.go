package mapper

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/norman/types/values"
)

type RevisionMapper struct {
	Field string
}

func (d RevisionMapper) FromInternal(data map[string]interface{}) {
	v, ok := data[d.Field]
	if !ok {
		return
	}

	m := convert.ToMapInterface(v)
	for k, v := range m {
		s := convert.ToMapInterface(v)
		parent := convert.ToString(s["parentService"])
		ver := convert.ToString(s["version"])
		if _, ok := m[parent]; parent != "" && ver != "" && ok {
			delete(s, "parentService")
			delete(s, "version")
			values.PutValue(m, s, parent, "revisions", ver)
			delete(m, k)
		}
	}
}

func (d RevisionMapper) ToInternal(data map[string]interface{}) error {
	v, ok := data[d.Field]
	if !ok {
		return nil
	}

	m := convert.ToMapInterface(v)
	for k, v := range m {
		s := convert.ToMapInterface(v)
		revs := convert.ToMapInterface(values.GetValueN(s, "revisions"))
		for ver, verser := range revs {
			mverser := convert.ToMapInterface(verser)
			if len(mverser) > 0 {
				mverser["parentService"] = k
				mverser["version"] = ver
			}
		}
		delete(s, "revisions")
	}

	return nil
}

func (d RevisionMapper) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return nil
}
