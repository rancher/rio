package configmap

import (
	"encoding/base64"

	"github.com/pkg/errors"
	"github.com/rancher/rio/pkg/deploy/stack/input"
	"github.com/rancher/rio/pkg/deploy/stack/output"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Populate(stack *input.Stack, output *output.Deployment) error {
	configMaps := map[string]*v1.ConfigMap{}

	for _, config := range stack.Configs {
		cfg := newConfig(config.Name, stack.Namespace, map[string]string{
			"rio.cattle.io/config":    config.Name,
			"rio.cattle.io/workspace": stack.Stack.Namespace,
			"rio.cattle.io/stack":     stack.Stack.Name,
		})
		if config.Spec.Encoded {
			bytes, err := base64.StdEncoding.DecodeString(config.Spec.Content)
			if err != nil {
				return errors.Wrapf(err, "failed to decode data for %s", config.Name)
			}
			cfg.BinaryData = map[string][]byte{
				"content": bytes,
			}
		} else {
			cfg.Data = map[string]string{
				"content": config.Spec.Content,
			}
		}
		configMaps[config.Name] = cfg
	}

	output.ConfigMaps = configMaps
	return nil
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
