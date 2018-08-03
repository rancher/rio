package config

import "fmt"

type IstioInjector struct {
	configFactory *ConfigFactory
}

func NewIstioInjector(cf *ConfigFactory) *IstioInjector {
	return &IstioInjector{
		configFactory: cf,
	}
}

func (i *IstioInjector) Inject(content []byte) ([]byte, error) {
	meshConfig, template := i.configFactory.TemplateAndConfig()
	if meshConfig == nil {
		return nil, fmt.Errorf("waiting on mesh configuration")
	}

	return Inject(content, template, meshConfig)
}
