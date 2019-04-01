package populate

import (
	"fmt"
	"strings"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func ToPVC(labels map[string]string, volume riov1.Volume) (*v1.PersistentVolumeClaim, error) {
	cfg := constructors.NewPersistentVolumeClaim(volume.Namespace, volume.Name, v1.PersistentVolumeClaim{})
	cfg.Labels = map[string]string{}
	for k, v := range labels {
		cfg.Labels[k] = v
	}

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

func Volume(volume *riov1.Volume, output *objectset.ObjectSet) error {
	return addVolume(volume, output)
}

func addVolume(volume *riov1.Volume, output *objectset.ObjectSet) error {
	if volume.Spec.Template {
		return nil
	}

	cfg, err := ToPVC(nil, *volume)
	if err != nil {
		return err
	}

	output.Add(cfg)
	return nil
}
