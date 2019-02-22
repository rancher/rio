package setup

import (
	"context"

	"github.com/rancher/norman/store/crd"
	"github.com/rancher/norman/store/proxy"
	normantypes "github.com/rancher/norman/types"
	buildv1alpha1 "github.com/rancher/rio/types/apis/build.knative.dev/v1alpha1"
	buildschema "github.com/rancher/rio/types/apis/build.knative.dev/v1alpha1/schema"
	"github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3"
	networkSchema "github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3/schema"
	autoscalev1 "github.com/rancher/rio/types/apis/rio-autoscale.cattle.io/v1"
	autoscaleSchema "github.com/rancher/rio/types/apis/rio-autoscale.cattle.io/v1/schema"
	webhookschema "github.com/rancher/rio/types/apis/webhookinator.rio.cattle.io/v1"
)

func Types(ctx context.Context, clientGetter proxy.ClientGetter, schemas *normantypes.Schemas) error {
	//server := norman.GetServer(ctx)
	factory := crd.NewFactoryFromClientGetter(clientGetter)
	// We create istio types so that our controllers don't error on first start
	_, err := factory.CreateCRDs(ctx, normantypes.DefaultStorageContext,
		networkSchema.Schemas.Schema(&networkSchema.Version, v1alpha3.GatewayGroupVersionKind.Kind),
		networkSchema.Schemas.Schema(&networkSchema.Version, v1alpha3.VirtualServiceGroupVersionKind.Kind),
		networkSchema.Schemas.Schema(&networkSchema.Version, v1alpha3.DestinationRuleGroupVersionKind.Kind),
		networkSchema.Schemas.Schema(&networkSchema.Version, v1alpha3.ServiceEntryGroupVersionKind.Kind),
		autoscaleSchema.Schemas.Schema(&autoscaleSchema.APIVersion, autoscalev1.ServiceScaleRecommendationGroupVersionKind.Kind),
		buildschema.Schemas.Schema(&buildschema.Version, buildv1alpha1.BuildGroupVersionKind.Kind),
		webhookschema.Schemas.Schema(&webhookschema.APIVersion, webhookschema.GitWebHookExecutionGroupVersionKind.Kind),
		webhookschema.Schemas.Schema(&webhookschema.APIVersion, webhookschema.GitWebHookReceiverGroupVersionKind.Kind),
	)
	if err != nil {
		return err
	}

	factory.BatchWait()

	return nil
}
