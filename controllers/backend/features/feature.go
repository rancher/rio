package features

import (
	"context"

	"github.com/rancher/rio/controllers/backend/features/letsencrypt"
	"github.com/rancher/rio/controllers/backend/features/monitoring"
	"github.com/rancher/rio/controllers/backend/features/nfs"
	"github.com/rancher/rio/types"
	projectv1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type featureController struct {
	features           projectv1.FeatureController
	publicdomainLister projectv1.PublicDomainClientCache
	stacks             v1.StackClient
}

func Register(ctx context.Context, rContext *types.Context) {
	f := featureController{
		features:           rContext.Global.Feature.Interface().Controller(),
		publicdomainLister: rContext.Global.PublicDomain.Cache(),
		stacks:             rContext.Rio.Stack,
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
	case "letsencrypt":
		w := letsencrypt.Wrapper{
			PublicdomainLister: f.publicdomainLister,
			Feature:            feature,
			Stacks:             f.stacks,
		}
		return feature, w.Reconcile()
	}
	return feature, nil
}
