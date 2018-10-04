package volume

import (
	"context"
	"reflect"
	"strings"

	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

func Register(ctx context.Context, rContext *types.Context) {
	vc := &volumeController{
		volumeLister: rContext.Rio.Volumes("").Controller().Lister(),
		volumes:      rContext.Rio.Volumes(""),
	}

	rContext.Core.PersistentVolumeClaims("").AddHandler("volume-controller", vc.sync)
	rContext.Core.PersistentVolumeClaims("").AddLifecycle("volume-template-controller", vc)
}

type volumeController struct {
	volumeLister v1beta1.VolumeLister
	volumes      v1beta1.VolumeInterface
}

func (v *volumeController) sync(key string, pvc *v1.PersistentVolumeClaim) error {
	if pvc == nil {
		return nil
	}

	_, ok := pvc.Labels["rio.cattle.io/workspace"]
	if !ok {
		return nil
	}

	name, ok := pvc.Labels["rio.cattle.io/volume"]
	if !ok {
		return nil
	}

	vol, err := v.volumeLister.Get(pvc.Namespace, name)
	if errors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	if reflect.DeepEqual(vol.Status.PVCStatus, &pvc.Status) {
		return nil
	}

	newVol := vol.DeepCopy()
	newVol.Status.PVCStatus = &pvc.Status
	_, err = v.volumes.Update(newVol)
	return err
}

func (v *volumeController) Create(pvc *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	if pvc == nil {
		return pvc, nil
	}

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

	volume := &v1beta1.Volume{}
	volume.Spec = volTemplate.Spec
	volume.Name = pvc.Name
	volume.Namespace = pvc.Namespace
	volume.Spec.Template = false
	volume.Labels = make(map[string]string)
	volume.Labels["rio.cattle.io/volume-template"] = volTemplate.Name
	if _, err = v.volumes.Create(volume); err != nil && !errors.IsAlreadyExists(err) {
		return pvc, err
	}
	return pvc, nil
}

func (v *volumeController) Updated(pvc *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	return pvc, nil
}

func (v *volumeController) Remove(pvc *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	return pvc, nil
}
