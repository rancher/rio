package feature

import (
	"context"

	"github.com/rancher/rio/modules/linkerd/controller/inject"
	"github.com/rancher/rio/modules/linkerd/pkg/injector"
	"github.com/rancher/rio/pkg/arch"
	"github.com/rancher/rio/pkg/config"
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
	enabled, err := shouldRunLinkerd(rContext)
	if err != nil {
		return err
	}
	cmClient := rContext.Core.Core().V1().ConfigMap()

	conf, err := config.GetConfig(rContext.Namespace, cmClient)
	if err != nil {
		return err
	}

	linkerdInstallStack := stack.NewSystemStack(
		rContext.Apply.WithCacheTypes(rContext.Batch.Batch().V1().Job()),
		rContext.Admin.Admin().V1().SystemStack(),
		rContext.Namespace,
		"linkerd-install",
	)

	feature := &features.FeatureController{
		FeatureName: "linkerd",
		FeatureSpec: features.FeatureSpec{
			Description: "linkerd service mesh",
			Enabled:     enabled,
		},
		Controllers: []features.ControllerRegister{
			inject.Register,
		},
		SystemStacks: []*stack.SystemStack{
			linkerdInstallStack,
		},
		FixedAnswers: map[string]string{
			"LINKERD_INSTALL_IMAGE": constants.LinkerdInstallImage,
			"NAMESPACE":             rContext.Namespace,
		},
		OnStart: func() error {
			injector.RegisterInjector()
			rContext.Rio.Rio().V1().Service().Enqueue("*", "*")
			return nil
		},
	}

	// only upgrade linkerd when configmap upgrade option is true
	if _, err := cmClient.Get(linkerdNamespace, linkerdConfigName, metav1.GetOptions{}); err == nil {
		if conf.Features["linkerd"].Options["upgrade"] == "true" {
			feature.FixedAnswers["LINKERD_UPGRADE"] = "\"TRUE\""
		} else {
			feature.SystemStacks = nil
		}
	}

	return feature.Register()
}

func shouldRunLinkerd(rContext *types.Context) (bool, error) {
	if !arch.IsAmd64() {
		return false, nil
	}
	var totalMemory int64
	nodes, err := rContext.Core.Core().V1().Node().List(metav1.ListOptions{})
	if err != nil {
		return false, err
	}
	for _, node := range nodes.Items {
		totalMemory += node.Status.Capacity.Memory().Value()
	}
	return totalMemory > 2147000000, nil
}
