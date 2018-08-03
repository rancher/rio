package deploy

import (
	"encoding/base64"

	"github.com/pkg/errors"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func configs(objects []runtime.Object, stack *StackResources, namespace string) (map[string]*v1beta1.Config, []runtime.Object, error) {
	configMaps := map[string]*v1beta1.Config{}

	for _, config := range stack.Configs {
		configMaps[config.Name] = config
		cfg := newConfig(config.Name, namespace, map[string]string{
			"rio.cattle.io/namespace": namespace,
		})
		if config.Spec.Encoded {
			bytes, err := base64.StdEncoding.DecodeString(config.Spec.Content)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "failed to decode data for %s", config.Name)
			}
			cfg.BinaryData = map[string][]byte{
				"content": bytes,
			}
		} else {
			cfg.Data = map[string]string{
				"content": config.Spec.Content,
			}
		}
		objects = append(objects, cfg)
	}

	return configMaps, objects, nil
}

func newConfig(name, namespace string, labels map[string]string) *v1.ConfigMap {
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: map[string]string{},
		},
	}
}
