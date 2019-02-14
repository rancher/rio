package setup

import (
	"context"

	"github.com/rancher/norman/api/builtin"
	"github.com/rancher/norman/pkg/subscribe"
	"github.com/rancher/norman/store/crd"
	"github.com/rancher/norman/store/proxy"
	normantypes "github.com/rancher/norman/types"
	"github.com/rancher/rio/api/config"
	"github.com/rancher/rio/api/meshnamed"
	"github.com/rancher/rio/api/named"
	"github.com/rancher/rio/api/pretty"
	"github.com/rancher/rio/api/publicdomain"
	"github.com/rancher/rio/api/resetstack"
	"github.com/rancher/rio/api/service"
	"github.com/rancher/rio/api/space"
	"github.com/rancher/rio/api/stack"
	buildv1alpha1 "github.com/rancher/rio/types/apis/build.knative.dev/v1alpha1"
	buildschema "github.com/rancher/rio/types/apis/build.knative.dev/v1alpha1/schema"
	"github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3"
	networkSchema "github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3/schema"
	projectSchema "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1/schema"
	autoscalev1 "github.com/rancher/rio/types/apis/rio-autoscale.cattle.io/v1"
	autoscaleSchema "github.com/rancher/rio/types/apis/rio-autoscale.cattle.io/v1/schema"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1/schema"
	projectclient "github.com/rancher/rio/types/client/project/v1"
	client "github.com/rancher/rio/types/client/rio/v1"
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
	)
	if err != nil {
		return err
	}

	factory.BatchWait()

	setupSpaces(ctx, clientGetter, schemas)
	setupNodes(ctx, clientGetter, schemas)
	setupPods(ctx, clientGetter, schemas)
	setupService(ctx, schemas)
	setupConfig(ctx, schemas)
	setupRoute(ctx, schemas)
	setupVolume(ctx, schemas)
	setupStacks(ctx, schemas)
	setupPublicDomain(ctx, schemas)
	setupExternalService(ctx, schemas)

	subscribe.Register(&builtin.Version, schemas)
	subscribe.Register(&schema.Version, schemas)
	subscribe.Register(&projectSchema.Version, schemas)

	return nil
}

func setupService(ctx context.Context, schemas *normantypes.Schemas) {
	s := schemas.Schema(&schema.Version, client.ServiceType)
	s.Formatter = pretty.Format
	s.InputFormatter = pretty.InputFormatter
	s.Store = resetstack.New(service.New(meshnamed.New(named.New(s.Store))))
}

func setupConfig(ctx context.Context, schemas *normantypes.Schemas) {
	s := schemas.Schema(&schema.Version, client.ConfigType)
	s.Store = resetstack.New(named.New(s.Store))
	s.ListHandler = config.ListHandler
}

func setupRoute(ctx context.Context, schemas *normantypes.Schemas) {
	s := schemas.Schema(&schema.Version, client.RouteSetType)
	s.Formatter = pretty.Format
	s.InputFormatter = pretty.InputFormatter
	s.Store = resetstack.New(meshnamed.New(named.New(s.Store)))
}

func setupVolume(ctx context.Context, schemas *normantypes.Schemas) {
	s := schemas.Schema(&schema.Version, client.VolumeType)
	s.Store = resetstack.New(named.New(s.Store))
}

func setupExternalService(ctx context.Context, schemas *normantypes.Schemas) {
	s := schemas.Schema(&schema.Version, client.ExternalServiceType)
	s.Store = resetstack.New(meshnamed.New(named.New(s.Store)))
}

func setupStacks(ctx context.Context, schemas *normantypes.Schemas) {
	s := schemas.Schema(&schema.Version, client.StackType)
	s.Formatter = pretty.Format
	s.ListHandler = stack.ListHandler
	s.Store = named.New(s.Store)
}

func setupNodes(ctx context.Context, clientGetter proxy.ClientGetter, schemas *normantypes.Schemas) {
	s := schemas.Schema(&projectSchema.Version, projectclient.NodeType)
	s.Store = proxy.NewProxyStore(ctx,
		clientGetter,
		normantypes.DefaultStorageContext,
		[]string{"/api"},
		"",
		"v1",
		"Node",
		"nodes")
}

func setupPods(ctx context.Context, clientGetter proxy.ClientGetter, schemas *normantypes.Schemas) {
	s := schemas.Schema(&projectSchema.Version, projectclient.PodType)
	s.Store = proxy.NewProxyStore(ctx,
		clientGetter,
		normantypes.DefaultStorageContext,
		[]string{"/api"},
		"",
		"v1",
		"Pod",
		"pods")
}

func setupSpaces(ctx context.Context, clientGetter proxy.ClientGetter, schemas *normantypes.Schemas) {
	s := schemas.Schema(&projectSchema.Version, projectclient.ProjectType)
	s.Store = proxy.NewProxyStore(ctx,
		clientGetter,
		normantypes.DefaultStorageContext,
		[]string{"/api"},
		"",
		"v1",
		"Namespace",
		"namespaces")
	s.Store = space.New(s.Store)
}

func setupPublicDomain(ctx context.Context, schemas *normantypes.Schemas) {
	s := schemas.Schema(&projectSchema.Version, projectclient.PublicDomainType)
	s.Store = publicdomain.New(named.New(s.Store))
}
