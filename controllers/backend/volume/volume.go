package volume

import (
	"context"
	"reflect"
	"strings"

	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	vc := &volumeController{
		volumeLister: rContext.Rio.Volume.Cache(),
		volumes:      rContext.Rio.Volume,
	}

	rContext.Core.PersistentVolumeClaim.OnChange(ctx, "volume-controller", vc.sync)
	rContext.Core.PersistentVolumeClaim.OnChange(ctx, "volume-template-controller", vc.syncTemplate)

	return nil
}

type volumeController struct {
	volumeLister v1beta1.VolumeClientCache
	volumes      v1beta1.VolumeClient
}

func (v *volumeController) sync(pvc *v1.PersistentVolumeClaim) (runtime.Object, error) {
	_, ok := pvc.Labels["rio.cattle.io/workspace"]
	if !ok {
		return nil, nil
	}

	name, ok := pvc.Labels["rio.cattle.io/volume"]
	if !ok {
		return nil, nil
	}

	vol, err := v.volumeLister.Get(pvc.Namespace, name)
	if errors.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	if reflect.DeepEqual(vol.Status.PVCStatus, &pvc.Status) {
		return nil, nil
	}

	newVol := vol.DeepCopy()
	newVol.Status.PVCStatus = &pvc.Status
	_, err = v.volumes.Update(newVol)
	return nil, err
}

func (v *volumeController) syncTemplate(pvc *v1.PersistentVolumeClaim) (runtime.Object, error) {
	if pvc.Labels["rio.cattle.io/use-templates"] == "" {
		return nil, nil
	}

	ns := pvc.Namespace
	serviceName := pvc.Labels["app"]
	templateName := strings.SplitN(pvc.Name, "-"+serviceName+"-", 2)
	volTemplate, err := v.volumeLister.Get(ns, templateName[0])
	if errors.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	volume := &v1beta1.Volume{}
	volume.Spec = volTemplate.Spec
	volume.Name = pvc.Name
	volume.Namespace = pvc.Namespace
	volume.Spec.Template = false
	volume.Labels = make(map[string]string)
	volume.Labels["rio.cattle.io/volume-template"] = volTemplate.Name
	if _, err := v.volumeLister.Get(ns, volume.Name); err != nil {
		if errors.IsNotFound(err) {
			if _, err = v.volumes.Create(volume); err != nil && !errors.IsAlreadyExists(err) {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return nil, nil
}

func (v *volumeController) Updated(pvc *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	return pvc, nil
}

func (v *volumeController) Remove(pvc *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	return pvc, nil
}
