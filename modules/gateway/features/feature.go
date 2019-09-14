package features

import (
	"context"
	"fmt"

	"github.com/rancher/rio/modules/gateway/controllers/app"
	"github.com/rancher/rio/modules/gateway/controllers/externalservice"
	"github.com/rancher/rio/modules/gateway/controllers/istio"
	"github.com/rancher/rio/modules/gateway/controllers/publicdomain"
	"github.com/rancher/rio/modules/gateway/controllers/routeset"
	"github.com/rancher/rio/modules/gateway/controllers/service"
	feature2 "github.com/rancher/rio/modules/linkerd/feature"
	projectv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/pkg/template/gotemplate"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/start"
)

func Register(ctx context.Context, rContext *types.Context) error {
	linkerdInstall, err := feature2.ConfigureLinkerdInstall(rContext)
	if err != nil {
		return err
	}

	var proxyInject []byte
	if constants.ServiceMeshMode == constants.ServiceMeshModeLinkerd {
		proxyInject, err = gotemplate.Apply([]byte(feature2.ProxyConfig), map[string]string{
			"CA_PEM":               string(linkerdInstall.Data["ca"]),
			"NAMESPACE":            rContext.Namespace,
			"PROXY_DESTINATION":    fmt.Sprintf("linkerd-destination.%s.svc.cluster.local:8086", rContext.Namespace),
			"IDENTITY_DESTINATION": fmt.Sprintf("linkerd-identity.%s.svc.cluster.local:8080", rContext.Namespace),
		})
		if err != nil {
			return err
		}
	}

	apply := rContext.Apply.WithCacheTypes(rContext.Rio.Rio().V1().Service(), rContext.Core.Core().V1().ConfigMap())
	feature := &features.FeatureController{
		FeatureName: "gateway",
		FeatureSpec: projectv1.FeatureSpec{
			Description: "Gateway service based on pilot and envoy",
			Enabled:     true,
		},
		SystemStacks: []*stack.SystemStack{
			stack.NewSystemStack(apply, rContext.Namespace, "gateway-crd"),
			stack.NewSystemStack(apply, rContext.Namespace, "gateway"),
		},
		Controllers: []features.ControllerRegister{
			externalservice.Register,
			istio.Register,
			routeset.Register,
			service.Register,
			app.Register,
			publicdomain.Register,
		},
		FixedAnswers: map[string]string{
			"HTTP_PORT":         constants.DefaultHTTPOpenPort,
			"HTTPS_PORT":        constants.DefaultHTTPSOpenPort,
			"TELEMETRY_ADDRESS": fmt.Sprintf("%s.%s.svc.cluster.local", constants.IstioTelemetry, rContext.Namespace),
			"NAMESPACE":         rContext.Namespace,
			"TAG":               constants.IstioVersion,
			"LINKERD_TAG":       constants.LinkerdVersion,
			"INSTALL_MODE":      constants.InstallMode,
			"PROXY_INJECT":      string(proxyInject),
		},
		OnStart: func(feature *projectv1.Feature) error {
			return start.All(ctx, 5,
				rContext.Global,
				rContext.Networking,
				rContext.K8sNetworking,
			)
		},
	}

	return feature.Register()
}
