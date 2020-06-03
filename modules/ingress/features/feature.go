package features

import (
	"context"

	config2 "github.com/rancher/rio/modules/ingress/controllers/config"
	"github.com/rancher/rio/modules/ingress/controllers/ingress"
	"github.com/rancher/rio/pkg/arch"
	"github.com/rancher/rio/pkg/config"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Register(ctx context.Context, rContext *types.Context) error {
	cm, err := rContext.Core.Core().V1().ConfigMap().Get(rContext.Namespace, config.ConfigName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	cfg, err := config.FromConfigMap(cm)
	if err != nil {
		return err
	}

	// if running in arm or gloo feature is not enabled, enable ingress
	enabled := false
	if !arch.IsAmd64() || (!featureEnabled(cfg, "gloo") && !featureEnabled(cfg, "istio")) {
		enabled = true
	}
	feature := &features.FeatureController{
		FeatureName: "ingress",
		FeatureSpec: features.FeatureSpec{
			Description: "Ingress as API gateway",
			Enabled:     enabled,
		},
		Controllers: []features.ControllerRegister{
			config2.Register,
			ingress.Register,
		},
	}
	return feature.Register()
}

func featureEnabled(cfg config.Config, name string) bool {
	if cfg.Features[name].Enabled != nil {
		return *cfg.Features[name].Enabled
	}
	return false
}
