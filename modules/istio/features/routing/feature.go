package routing

import (
	"context"
	"encoding/base64"

	projectv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/start"
)

func Register(ctx context.Context, rContext *types.Context) error {
	apply := rContext.Apply.WithCacheTypes(rContext.Rio.Rio().V1().Service(), rContext.Core.Core().V1().ConfigMap())

	systemStacks := []*stack.SystemStack{
		stack.NewSystemStack(apply, rContext.Namespace, "istio-mesh"),
		stack.NewSystemStack(apply, rContext.Namespace, "istio-crd"),
		stack.NewSystemStack(apply, rContext.Namespace, "istio"),
	}
	if !constants.DisableGrafana {
		systemStacks = append(systemStacks, stack.NewSystemStack(apply, rContext.Namespace, "istio-grafana"))
	}

	disableKiali := ""
	if constants.DisableKiali {
		disableKiali = "true"
	}
	disablePrometheus := ""
	if constants.DisablePrometheus {
		disablePrometheus = "true"
	}
	disableMixer := ""
	if constants.DisableMixer {
		disableMixer = "true"
	}
	feature := &features.FeatureController{
		FeatureName: "istio",
		FeatureSpec: projectv1.FeatureSpec{
			Description: "Istio service mesh",
			Enabled:     constants.ServiceMeshMode == constants.ServiceMeshModeIstio,
			Answers: map[string]string{
				"DISABLE_KIALI":      disableKiali,
				"DISABLE_MIXER":      disableMixer,
				"DISABLE_PROMETHEUS": disablePrometheus,
				"KIALI_USERNAME":     base64.StdEncoding.EncodeToString([]byte("admin")),
				"KIALI_PASSPHRASE":   base64.StdEncoding.EncodeToString([]byte("admin")),
			},
		},
		SystemStacks: systemStacks,
		Controllers:  []features.ControllerRegister{},
		FixedAnswers: map[string]string{
			"NAMESPACE": rContext.Namespace,
			"TAG":       constants.IstioVersion,
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
