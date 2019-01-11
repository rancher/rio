package pvc

import (
	"context"
	"reflect"
	"strings"

	"github.com/rancher/rio/types"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	vc := &pvcHandler{
		volumeLister: rContext.Rio.Volume.Cache(),
		volumes:      rContext.Rio.Volume,
	}

	rContext.Core.PersistentVolumeClaim.OnChange(ctx, "stack-pvc-controller", vc.sync)
	rContext.Core.PersistentVolumeClaim.OnChange(ctx, "stack-pvc-template-controller", vc.syncTemplate)

	return nil
}

type pvcHandler struct {
	volumeLister riov1.VolumeClientCache
	volumes      riov1.VolumeClient
}

func (v *pvcHandler) sync(pvc *v1.PersistentVolumeClaim) (runtime.Object, error) {
	_, ok := pvc.Labels["rio.cattle.io/project"]
	if !ok {
		return pvc, nil
	}

	name, ok := pvc.Labels["rio.cattle.io/volume"]
	if !ok {
		return pvc, nil
	}

	vol, err := v.volumeLister.Get(pvc.Namespace, name)
	if errors.IsNotFound(err) {
		return pvc, nil
	} else if err != nil {
		return pvc, err
	}

	if reflect.DeepEqual(vol.Status.PVCStatus, &pvc.Status) {
		return pvc, nil
	}

	newVol := vol.DeepCopy()
	newVol.Status.PVCStatus = &pvc.Status
	_, err = v.volumes.Update(newVol)
	return pvc, err
}

func (v *pvcHandler) syncTemplate(pvc *v1.PersistentVolumeClaim) (runtime.Object, error) {
	if pvc.Labels["rio.cattle.io/use-templates"] == "" {
		return pvc, nil
	}

	ns := pvc.Namespace
	serviceName := pvc.Labels["app"]
	templateName := strings.SplitN(pvc.Name, "-"+serviceName+"-", 2)
	volTemplate, err := v.volumeLister.Get(ns, templateName[0])
	if errors.IsNotFound(err) {
		return pvc, nil
	} else if err != nil {
		return pvc, err
	}

	volume := &riov1.Volume{}
	volume.Spec = volTemplate.Spec
	volume.Name = pvc.Name
	volume.Namespace = pvc.Namespace
	volume.Spec.Template = false
	volume.Labels = make(map[string]string)
	volume.Labels["rio.cattle.io/volume-template"] = volTemplate.Name
	if _, err := v.volumeLister.Get(ns, volume.Name); err != nil {
		if errors.IsNotFound(err) {
			if _, err = v.volumes.Create(volume); err != nil && !errors.IsAlreadyExists(err) {
				return pvc, err
			}
		} else {
			return pvc, err
		}
	}
	return pvc, nil
}
