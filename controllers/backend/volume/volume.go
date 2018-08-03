package volume

import (
	"context"

	"reflect"

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
}

type volumeController struct {
	volumeLister v1beta1.VolumeLister
	volumes      v1beta1.VolumeInterface
}

func (v *volumeController) sync(key string, pvc *v1.PersistentVolumeClaim) error {
	if pvc == nil {
		return nil
	}

	ns, ok := pvc.Labels["rio.cattle.io/namespace"]
	if !ok {
		return nil
	}

	name, ok := pvc.Labels["rio.cattle.io/volume"]
	if !ok {
		return nil
	}

	vol, err := v.volumeLister.Get(ns, name)
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
