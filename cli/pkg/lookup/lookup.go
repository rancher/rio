package lookup

import (
	"fmt"

	"github.com/rancher/rio/cli/pkg/types"
	"github.com/rancher/rio/cli/pkg/up/questions"
)

type ClientLookup interface {
	ByID(id, typeName string) (types.Resource, error)
}

func Lookup(ctx ClientLookup, name string, typeNames ...string) (types.Resource, error) {
	var result []types.Resource
	for _, schemaType := range typeNames {
		r, err := ctx.ByID(name, schemaType)
		if err == nil && r.Type != "" {
			result = append(result, types.Resource{
				Name:      r.Name,
				Namespace: r.Namespace,
				Type:      r.Type,
			})
		}
	}

	if len(result) == 0 {
		return types.Resource{}, fmt.Errorf("not found (types=%v): %s", typeNames, name)
	}

	if len(result) == 1 {
		return result[0], nil
	}

	msg := fmt.Sprintf("Choose resource for %s:\n", name)
	var options []string

	for i, r := range result {
		msg := fmt.Sprintf("[%d] type=%s %s\n", i+1, r.Type, r.Name)
		options = append(options, msg)
	}

	num, err := questions.PromptOptions(msg, -1, options...)
	if err != nil {
		return types.Resource{}, err
	}
	return result[num], nil
}
