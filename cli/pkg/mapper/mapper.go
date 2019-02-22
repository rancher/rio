package mapper

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/types/mapper"
)

type Wrapper struct {
	mapper []types.Mapper
}

var GenericStatusMapper = Wrapper{
	mapper: []types.Mapper{
		mapper.Status{},
	},
}

func (w *Wrapper) AddMapper(m ...types.Mapper) {
	w.mapper = append(w.mapper, m...)
}

func (w *Wrapper) FormatState(data map[string]interface{}) (string, error) {
	for _, m := range w.mapper {
		m.FromInternal(data)
	}
	return data["state"].(string), nil
}

func (w *Wrapper) FormatTransitionMessage(data map[string]interface{}) (string, error) {
	for _, m := range w.mapper {
		m.FromInternal(data)
	}
	return data["transitioningMessage"].(string), nil
}
