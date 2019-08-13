package publicdomain

import (
	"context"

	"github.com/rancher/rio/modules/istio/controllers/publicdomain/populate"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "routing-publicdomain", rContext.Global.Admin().V1().PublicDomain())
	c.Apply = c.Apply.WithCacheTypes(rContext.Networking.Networking().V1alpha3().DestinationRule(),
		rContext.K8sNetworking.Networking().V1beta1().Ingress())
	p := populator{
		systemNamespace: rContext.Namespace,
	}
	c.Populator = p.populate
	return nil
}

type populator struct {
	systemNamespace string
}

func (p populator) populate(obj runtime.Object, ns *corev1.Namespace, os *objectset.ObjectSet) error {
	pd := obj.(*adminv1.PublicDomain)
	if constants.InstallMode == constants.InstallModeIngress {
		populate.Ingress(p.systemNamespace, pd, os)
	}
	return populate.DestionationRule(obj.(*adminv1.PublicDomain), p.systemNamespace, os)
}
