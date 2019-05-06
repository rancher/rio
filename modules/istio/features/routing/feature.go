package routing

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rancher/rio/pkg/settings"

	"github.com/rancher/rio/modules/istio/controllers/app"
	"github.com/rancher/rio/modules/istio/controllers/externalservice"
	"github.com/rancher/rio/modules/istio/controllers/istio"
	"github.com/rancher/rio/modules/istio/controllers/routeset"
	"github.com/rancher/rio/modules/istio/controllers/service"
	projectv1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/systemstack"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	apply := rContext.Apply.WithCacheTypes(rContext.Rio.Rio().V1().Service(), rContext.Core.Core().V1().ConfigMap())
	ports := []string{
		fmt.Sprintf("%v:%v,http2", settings.DefaultHTTPOpenPort, settings.DefaultHTTPOpenPort),
		fmt.Sprintf("%v:%v,https", settings.DefaultHTTPSOpenPort, settings.DefaultHTTPSOpenPort),
	}

	portStr, err := json.Marshal(&ports)
	if err != nil {
		return err
	}
	feature := &features.FeatureController{
		FeatureName: "istio",
		FeatureSpec: projectv1.FeatureSpec{
			Description: "Service routing using Istio",
			Enabled:     true,
		},
		SystemStacks: []*systemstack.SystemStack{
			systemstack.NewStack(apply, rContext.Namespace, "mesh", true),
			systemstack.NewStack(apply, rContext.Namespace, "istio", true),
		},
		Controllers: []features.ControllerRegister{
			externalservice.Register,
			istio.Register,
			routeset.Register,
			service.Register,
			app.Register,
		},
		FixedAnswers: map[string]string{
			"PORTS":             string(portStr),
			"TELEMETRY_ADDRESS": fmt.Sprintf("%s.%s.svc.cluster.local", settings.IstioTelemetry, rContext.Namespace),
			"NAMESPACE":         rContext.Namespace,
		},
	}

	return feature.Register()
}
