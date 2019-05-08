package config

import (
	"bytes"

	"github.com/rancher/wrangler/pkg/yaml"
	"istio.io/api/mesh/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

type IstioInjector struct {
	meshConfig *v1alpha1.MeshConfig
	template   string
}

func NewIstioInjector(meshConfig *v1alpha1.MeshConfig, template string) *IstioInjector {
	return &IstioInjector{
		meshConfig: meshConfig,
		template:   template,
	}
}

func (i *IstioInjector) Inject(objs []runtime.Object) ([]runtime.Object, error) {
	meshConfig, template := i.meshConfig, i.template

	content, err := yaml.ToBytes(objs)
	if err != nil {
		return nil, err
	}

	output, err := Inject(content, template, meshConfig)
	if err != nil {
		return nil, err
	}

	return yaml.ToObjects(bytes.NewBuffer(output))
}
