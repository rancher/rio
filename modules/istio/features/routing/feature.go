package routing

import (
	"context"
	"fmt"

	"github.com/rancher/rio/modules/istio/controllers/app"
	"github.com/rancher/rio/modules/istio/controllers/externalservice"
	"github.com/rancher/rio/modules/istio/controllers/istio"
	"github.com/rancher/rio/modules/istio/controllers/publicdomain"
	"github.com/rancher/rio/modules/istio/controllers/routeset"
	"github.com/rancher/rio/modules/istio/controllers/service"
	projectv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	apply := rContext.Apply.WithCacheTypes(rContext.Rio.Rio().V1().Service(), rContext.Core.Core().V1().ConfigMap())
	feature := &features.FeatureController{
		FeatureName: "istio",
		FeatureSpec: projectv1.FeatureSpec{
			Description: "Service routing using Istio",
			Enabled:     !constants.DisableIstio,
		},
		SystemStacks: []*stack.SystemStack{
			stack.NewSystemStack(apply, rContext.Namespace, "mesh"),
			stack.NewSystemStack(apply, rContext.Namespace, "istio"),
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
			"TAG":               "1.1.3",
			"INSTALL_MODE":      constants.InstallMode,
		},
	}

	return feature.Register()
}
