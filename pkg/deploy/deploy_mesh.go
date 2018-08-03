package deploy

import (
	"github.com/rancher/rio/pkg/apply"
)

func DeployMesh(namespace string, stack *StackResources) error {
	objects, err := IstioObjects(namespace, stack)
	if err != nil {
		return err
	}

	namespaced, global, err := splitObjects(objects)
	if err != nil {
		return err
	}

	if len(global) > 0 {
		if err := apply.Apply(global, "stackdeploy-mesh-global-"+namespace, 0); err != nil {
			return err
		}
	}

	return apply.Apply(namespaced, "stackdeploy-mesh-"+namespace, 0)
}
