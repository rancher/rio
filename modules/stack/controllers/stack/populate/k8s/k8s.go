package k8s

import (
	"bytes"

	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/objectset"
	"github.com/rancher/wrangler/pkg/yaml"
	"k8s.io/apimachinery/pkg/api/meta"
)

func Populate(stack *v1.Stack, internalStack *v1.StackFile, output *objectset.ObjectSet) error {
	if internalStack.Kubernetes.Manifest != "" {
		err := readManifest("", internalStack.Kubernetes.Manifest, output)
		if err != nil {
			return err
		}
	}

	if internalStack.Kubernetes.NamespacedManifest != "" {
		err := readManifest(stack.Name, internalStack.Kubernetes.NamespacedManifest, output)
		if err != nil {
			return err
		}
	}

	return nil
}

func readManifest(namespace, content string, output *objectset.ObjectSet) error {
	objs, err := yaml.ToObjects(bytes.NewBufferString(content))
	if err != nil {
		return err
	}

	for _, obj := range objs {
		if namespace != "" {
			metadata, err := meta.Accessor(obj)
			if err != nil {
				return err
			}
			metadata.SetNamespace(namespace)
		}
		output.Add(obj)
	}

	return nil
}
