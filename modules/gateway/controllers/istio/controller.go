package istio

import (
	"context"
	"fmt"

	"github.com/rancher/rio/modules/gateway/controllers/istio/populate"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	adminv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/objectset"
	"github.com/rancher/wrangler/pkg/trigger"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	istioStack = "istio-stack"
)

var (
	addressTypes = []v1.NodeAddressType{
		v1.NodeExternalIP,
		v1.NodeInternalIP,
	}

	evalTrigger trigger.Trigger
)

func Register(ctx context.Context, rContext *types.Context) error {
	s := &istioDeployController{
		namespace: rContext.Namespace,
		gatewayApply: rContext.Apply.WithSetID(istioStack).WithStrictCaching().
			WithCacheTypes(rContext.Networking.Networking().V1alpha3().Gateway()),
		publicDomainCache: rContext.Global.Admin().V1().PublicDomain().Cache(),
	}

	rContext.Global.Admin().V1().ClusterDomain().OnChange(ctx, "clusterdomain-gateway", s.syncGateway)

	return nil
}

type istioDeployController struct {
	namespace         string
	gatewayApply      apply.Apply
	publicDomainCache adminv1controller.PublicDomainCache
}

func (i istioDeployController) syncGateway(key string, obj *adminv1.ClusterDomain) (*adminv1.ClusterDomain, error) {
	if obj == nil || obj.DeletionTimestamp != nil || obj.Name != constants.ClusterDomainName {
		return obj, nil
	}

	os := objectset.NewObjectSet()
	domain := ""
	if obj.Status.ClusterDomain != "" {
		domain = fmt.Sprintf("*.%s", obj.Status.ClusterDomain)
	}

	publicdomains, err := i.publicDomainCache.List("", labels.NewSelector())
	if err != nil {
		return obj, err
	}
	populate.Gateway(i.namespace, domain, obj.Spec.SecretRef.Name, publicdomains, os)

	return obj, i.gatewayApply.Apply(os)
}
