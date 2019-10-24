package setting

import (
	"context"

	"github.com/rancher/rio/pkg/config"
	gloov1controller "github.com/rancher/rio/pkg/generated/controllers/gloo.solo.io/v1"
	"github.com/rancher/rio/types"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1/kube/apis/gloo.solo.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Register(ctx context.Context, rContext *types.Context) error {
	cm, err := rContext.Core.Core().V1().ConfigMap().Get(rContext.Namespace, config.ConfigName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	cf, err := config.FromConfigMap(cm)
	if err != nil {
		return err
	}
	h := handler{
		systemNamespace: rContext.Namespace,
		gloo:            rContext.Gloo.Gloo().V1().Settings(),
		config:          cf,
	}
	rContext.Gloo.Gloo().V1().Settings().OnChange(ctx, "gloo-linkerd", h.onChange)
	return nil
}

type handler struct {
	systemNamespace string
	gloo            gloov1controller.SettingsClient
	config          config.Config
}

func (h handler) onChange(key string, settings *gloov1.Settings) (*gloov1.Settings, error) {
	if settings == nil || settings.DeletionTimestamp != nil {
		return settings, nil
	}

	if h.config.Features["linkerd"].Enabled != nil && !*h.config.Features["linkerd"].Enabled {
		return settings, nil
	}

	if settings.Namespace == h.systemNamespace && settings.Name == "default" {
		dp := settings.DeepCopy()
		dp.Spec.Linkerd = true
		return h.gloo.Update(settings)
	}

	return settings, nil
}
