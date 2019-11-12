package feature

import (
	"context"

	"github.com/rancher/rio/modules/linkerd/controller/inject"
	"github.com/rancher/rio/modules/linkerd/controller/prometheus"
	"github.com/rancher/rio/modules/linkerd/pkg/injector"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	linkerdNamespace  = "linkerd"
	linkerdConfigName = "linkerd-config"
)

func Register(ctx context.Context, rContext *types.Context) error {
	cmClient := rContext.Core.Core().V1().ConfigMap()
	linkerdUpgrade := ""
	if _, err := cmClient.Get(linkerdNamespace, linkerdConfigName, metav1.GetOptions{}); err == nil {
		linkerdUpgrade = "\"TRUE\""
	}
	apply := rContext.Apply.WithCacheTypes(rContext.Batch.Batch().V1().Job())
	feature := &features.FeatureController{
		FeatureName: "linkerd",
		FeatureSpec: features.FeatureSpec{
			Description: "linkerd service mesh",
			Enabled:     true,
		},
		SystemStacks: []*stack.SystemStack{
			stack.NewSystemStack(apply, rContext.Admin.Admin().V1().SystemStack(), rContext.Namespace, "linkerd-install"),
		},
		Controllers: []features.ControllerRegister{
			inject.Register,
			prometheus.Register,
		},
		FixedAnswers: map[string]string{
			"LINKERD_UPGRADE":       linkerdUpgrade,
			"LINKERD_INSTALL_IMAGE": constants.LinkerdInstallImage,
			"NAMESPACE":             rContext.Namespace,
		},
		OnStart: func() error {
			injector.RegisterInjector()
			rContext.Rio.Rio().V1().Service().Enqueue("*", "*")
			return nil
		},
	}
	return feature.Register()
}
