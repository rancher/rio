package features

import (
	"context"

	"github.com/rancher/rio/controllers/backend/features/monitoring"
	"github.com/rancher/rio/controllers/backend/features/nfs"
	"github.com/rancher/rio/types"
	projectv1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type featureController struct {
	features projectv1.FeatureController
}

func Register(ctx context.Context, rContext *types.Context) {
	f := featureController{
		features: rContext.Global.Feature.Interface().Controller(),
	}
	rContext.Global.Feature.Interface().AddHandler(ctx, "feature", f.sync)
}

func (f featureController) sync(key string, feature *projectv1.Feature) (runtime.Object, error) {
	if key == "" || feature == nil {
		return feature, nil
	}
	switch key {
	case "nfs":
		return feature, nfs.Reconcile(feature)
	case "monitoring":
		return feature, monitoring.Reconcile(feature)
	}
	return feature, nil
}
