package populate

import (
	"encoding/base64"

	"github.com/pkg/errors"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
)

func Config(stack *riov1.Stack, config *riov1.Config, os *objectset.ObjectSet) error {
	os.AddInput(stack, config)
	return addConfig(config, stack, os)
}

func addConfig(config *riov1.Config, stack *riov1.Stack, output *objectset.ObjectSet) error {
	cfg := constructors.NewConfigMap(stack.Name, config.Name, v1.ConfigMap{})
	cfg.Annotations = map[string]string{
		"rio.cattle.io/config":  config.Name,
		"rio.cattle.io/project": stack.Namespace,
		"rio.cattle.io/stack":   stack.Name,
	}

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

	output.Add(cfg)
	return nil
}
