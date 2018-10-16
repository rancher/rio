package volume

import (
	"fmt"
	"strings"

	"github.com/rancher/rio/pkg/deploy/stack/input"
	"github.com/rancher/rio/pkg/deploy/stack/output"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ToPVC(namespace string, labels map[string]string, volume v1beta1.Volume) (*v1.PersistentVolumeClaim, error) {
	cfg := newVolume(volume.Name, namespace, labels)
	if volume.Spec.Driver != "" && volume.Spec.Driver != "default" {
		cfg.Spec.StorageClassName = &volume.Spec.Driver
	}

	size := volume.Spec.SizeInGB
	if size == 0 {
		size = 10
	}

	q, err := resource.ParseQuantity(fmt.Sprintf("%dGi", size))
	if err != nil {
		return nil, fmt.Errorf("failed to parse size [%d] on volume %s", size, volume.Name)
	}

	switch strings.ToLower(volume.Spec.AccessMode) {
	case "readwritemany":
		cfg.Spec.AccessModes = []v1.PersistentVolumeAccessMode{
			v1.ReadWriteMany,
		}
	case "readonlymany":
		cfg.Spec.AccessModes = []v1.PersistentVolumeAccessMode{
			v1.ReadOnlyMany,
		}
	default:
		cfg.Spec.AccessModes = []v1.PersistentVolumeAccessMode{
			v1.ReadWriteOnce,
		}
	}

	cfg.Spec.Resources.Requests = v1.ResourceList{
		v1.ResourceStorage: q,
	}

	return cfg, nil
}

func Populate(stack *input.Stack, output *output.Deployment) error {
	result := map[string]*v1.PersistentVolumeClaim{}
	for _, volume := range stack.Volumes {
		if volume.Spec.Template {
			continue
		}

		labels := map[string]string{
			"rio.cattle.io/stack":     stack.Stack.Name,
			"rio.cattle.io/workspace": stack.Stack.Namespace,
			"rio.cattle.io/volume":    volume.Name,
		}

		cfg, err := ToPVC(stack.Namespace, labels, *volume)
		if err != nil {
			return err
		}

		result[cfg.Name] = cfg
	}

	output.PersistentVolumeClaims = result
	return nil
}

func newVolume(name, namespace string, labels map[string]string) *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
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
