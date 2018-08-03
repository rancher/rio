package config

import (
	"sync"

	"github.com/rancher/rio/cli/pkg/kv"
	"github.com/rancher/types/apis/core/v1"
	"istio.io/api/mesh/v1alpha1"
	metav1 "k8s.io/api/core/v1"
)

type ConfigFactory struct {
	sync.Mutex

	configMapNamespace string
	configMapName      string
	configMapKey       string
	template           string
	meshConfig         *v1alpha1.MeshConfig
}

func NewConfigFactory(configMap v1.ConfigMapInterface, configMapNamespace, configMapName, configMapKey string) *ConfigFactory {
	f := &ConfigFactory{
		configMapNamespace: configMapNamespace,
		configMapName:      configMapName,
		configMapKey:       configMapKey,
	}
	configMap.Controller().AddHandler("istio-config-cache", f.sync)
	return f
}

func (c *ConfigFactory) sync(key string, cm *metav1.ConfigMap) error {
	ns, name := kv.Split(key, "/")
	if ns != c.configMapNamespace && name != c.configMapName {
		return nil
	}

	if cm == nil {
		c.Lock()
		c.template = ""
		c.meshConfig = nil
		c.Unlock()
		return nil
	}

	val, ok := cm.Data[c.configMapKey]
	if !ok {
		return nil
	}

	meshConfig, template, err := ConfigAndTemplate(val)
	if err != nil {
		return err
	}

	c.Lock()
	c.template = template
	c.meshConfig = meshConfig
	c.Unlock()

	return nil
}

func (c *ConfigFactory) TemplateAndConfig() (*v1alpha1.MeshConfig, string) {
	c.Lock()
	defer c.Unlock()

	return c.meshConfig, c.template
}
