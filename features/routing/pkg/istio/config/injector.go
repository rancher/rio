package config

import (
	"fmt"

	"github.com/rancher/norman/pkg/objectset/injectors"
	"k8s.io/apimachinery/pkg/runtime"
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
		return nil, fmt.Errorf("waiting on mesh configuration")
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
