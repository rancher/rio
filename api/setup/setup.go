package setup

import (
	"context"

	"github.com/rancher/norman/api/builtin"
	"github.com/rancher/norman/pkg/subscribe"
	"github.com/rancher/norman/store/crd"
	"github.com/rancher/norman/store/proxy"
	normantypes "github.com/rancher/norman/types"
	"github.com/rancher/rio/api/config"
	"github.com/rancher/rio/api/defaults"
	"github.com/rancher/rio/api/named"
	"github.com/rancher/rio/api/pretty"
	"github.com/rancher/rio/api/resetstack"
	"github.com/rancher/rio/api/space"
	"github.com/rancher/rio/api/stack"
	"github.com/rancher/rio/types"
	networkSchema "github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3/schema"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1/schema"
	spaceSchema "github.com/rancher/rio/types/apis/space.cattle.io/v1beta1/schema"
	networkClient "github.com/rancher/rio/types/client/networking/v1alpha3"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	spaceClient "github.com/rancher/rio/types/client/space/v1beta1"
)

func SetupTypes(ctx context.Context, context *types.Context) error {
	factory := crd.NewFactoryFromClientGetter(context.ClientGetter)
	// We create istio types so that our controllers don't error on first start
	factory.CreateCRDs(ctx, normantypes.DefaultStorageContext,
		networkSchema.Schemas.Schema(&networkSchema.Version, networkClient.GatewayType),
		networkSchema.Schemas.Schema(&networkSchema.Version, networkClient.VirtualServiceType))
	factory.BatchCreateCRDs(ctx, normantypes.DefaultStorageContext, context.Schemas,
		&schema.Version,
		client.ServiceType,
		client.ConfigType,
		client.RouteSetType,
		client.VolumeType,
		client.StackType)
	factory.BatchCreateCRDs(ctx, normantypes.DefaultStorageContext, context.Schemas,
		&spaceSchema.Version,
		spaceClient.ListenConfigType)
	factory.BatchWait()

	setupSpaces(ctx, factory.ClientGetter, context)
	setupNodes(ctx, factory.ClientGetter, context)
	setupPods(ctx, factory.ClientGetter, context)
	setupService(ctx, context)
	setupConfig(ctx, context)
	setupRoute(ctx, context)
	setupVolume(ctx, context)
	setupStacks(ctx, context)

	subscribe.Register(&builtin.Version, context.Schemas)
	subscribe.Register(&schema.Version, context.Schemas)
	subscribe.Register(&spaceSchema.Version, context.Schemas)

	return nil
}

func setupService(ctx context.Context, rContext *types.Context) {
	s := rContext.Schemas.Schema(&schema.Version, client.ServiceType)
	s.Formatter = pretty.Format
	s.InputFormatter = pretty.InputFormatter
	s.Store = &defaults.DefaultStatusStore{
		Store: resetstack.New(named.New(s.Store)),
		Default: v1beta1.ServiceStatus{
			Conditions: []v1beta1.Condition{
				{
					Type:   "Pending",
					Status: "Unknown",
				},
			},
		},
	}
}

func setupConfig(ctx context.Context, rContext *types.Context) {
	s := rContext.Schemas.Schema(&schema.Version, client.ConfigType)
	s.Store = resetstack.New(named.New(s.Store))
	s.ListHandler = config.ListHandler
}

func setupRoute(ctx context.Context, rContext *types.Context) {
	s := rContext.Schemas.Schema(&schema.Version, client.RouteSetType)
	s.Formatter = pretty.Format
	s.InputFormatter = pretty.InputFormatter
	s.Store = resetstack.New(named.New(s.Store))
}

func setupVolume(ctx context.Context, rContext *types.Context) {
	s := rContext.Schemas.Schema(&schema.Version, client.VolumeType)
	s.Store = resetstack.New(named.New(s.Store))
}

func setupStacks(ctx context.Context, rContext *types.Context) {
	s := rContext.Schemas.Schema(&schema.Version, client.StackType)
	s.Formatter = pretty.Format
	s.ListHandler = stack.ListHandler
	s.Store = named.New(s.Store)
}

func setupNodes(ctx context.Context, clientGetter proxy.ClientGetter, rContext *types.Context) {
	s := rContext.Schemas.Schema(&spaceSchema.Version, spaceClient.NodeType)
	s.Store = proxy.NewProxyStore(ctx,
		clientGetter,
		normantypes.DefaultStorageContext,
		[]string{"/api"},
		"",
		"v1",
		"Node",
		"nodes")
}

func setupPods(ctx context.Context, clientGetter proxy.ClientGetter, rContext *types.Context) {
	s := rContext.Schemas.Schema(&spaceSchema.Version, spaceClient.PodType)
	s.Store = proxy.NewProxyStore(ctx,
		clientGetter,
		normantypes.DefaultStorageContext,
		[]string{"/api"},
		"",
		"v1",
		"Pod",
		"pods")
}

func setupSpaces(ctx context.Context, clientGetter proxy.ClientGetter, rContext *types.Context) {
	s := rContext.Schemas.Schema(&spaceSchema.Version, spaceClient.SpaceType)
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
