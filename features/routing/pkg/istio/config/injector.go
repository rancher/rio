package config

import (
	"github.com/pkg/errors"
	"github.com/rancher/norman/pkg/objectset/injectors"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type IstioInjector struct {
	configFactory *Factory
}

func NewIstioInjector(cf *Factory) *IstioInjector {
	return &IstioInjector{
		configFactory: cf,
	}
}

func (i *IstioInjector) Inject(objs []runtime.Object) ([]runtime.Object, error) {
	meshConfig, template := i.configFactory.TemplateAndConfig()
	if meshConfig == nil {
		return nil, errors2.NewConflict(schema.GroupResource{}, "", errors.New("waiting on mesh configuration"))
	}

	content, err := injectors.ToBytes(objs)
	if err != nil {
		return nil, err
	}

	output, err := Inject(content, template, meshConfig)
	if err != nil {
		return nil, err
	}

	return injectors.FromBytes(output)
}
