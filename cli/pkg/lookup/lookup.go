package lookup

import (
	"fmt"

	"github.com/rancher/rio/cli/pkg/types"
	"github.com/rancher/rio/cli/pkg/up/questions"
	"k8s.io/apimachinery/pkg/api/errors"
)

type ClientLookup interface {
	ByID(namespace, name, typeName string) (types.Resource, error)
	GetDefaultStackName() string
}

func getValidNameTypes(name string) (ret []nameType) {
	for _, nameType := range nameTypes {
		if nameType.Matches(name) {
			ret = append(ret, nameType)
		}
	}

	return
}

func Lookup(ctx ClientLookup, name string, typeNames ...string) (types.Resource, error) {
	resources, err := LookupAll(ctx, name, typeNames...)
	if err != nil {
		return types.Resource{}, err
	}

	if len(resources) == 0 {
		return types.Resource{}, fmt.Errorf("not found (types=%v): %s", typeNames, name)
	}

	if len(resources) == 1 {
		return resources[0], nil
	}

	msg := fmt.Sprintf("Choose resource for %s:\n", name)
	var options []string

	for i, r := range resources {
		msg := fmt.Sprintf("[%d] type=%s %s\n", i+1, r.Type, r.Name)
		if r.Namespace != "" {
			msg = fmt.Sprintf("[%d] type=%s %s/%s\n", i+1, r.Type, r.Namespace, r.Name)
		}
		options = append(options, msg)
	}

	num, err := questions.PromptOptions(msg, -1, options...)
	if err != nil {
		return types.Resource{}, err
	}
	return resources[num], nil
}

func LookupAll(ctx ClientLookup, name string, typeNames ...string) ([]types.Resource, error) {
	var (
		resources []types.Resource
		nameTypes = getValidNameTypes(name)
	)

outer:
	for _, typeName := range typeNames {
		for _, nameType := range nameTypes {
			r, err := nameType.Lookup(ctx, name, typeName)
			if errors.IsNotFound(err) {
				continue
			} else if err != nil {
				return nil, err
			}
			resources = append(resources, r)
			continue outer
		}
	}

	return resources, nil
}
