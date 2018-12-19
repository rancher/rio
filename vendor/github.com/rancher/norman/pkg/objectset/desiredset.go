package objectset

import (
	"github.com/rancher/norman/pkg/objectset/injectors"
	"github.com/rancher/norman/types"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type DesiredSet struct {
	setID       string
	objs        *ObjectSet
	codeVersion string
	clients     map[schema.GroupVersionKind]Client
	owner       runtime.Object
	injectors   []injectors.ConfigInjector
	errs        []error
}

func (o *DesiredSet) AddInjector(inj injectors.ConfigInjector) {
	o.injectors = append(o.injectors, inj)
}

func (o *DesiredSet) err(err error) error {
	o.errs = append(o.errs, err)
	return o.Err()
}

func (o *DesiredSet) Err() error {
	return types.NewErrors(append(o.objs.errs, o.errs...)...)
}
