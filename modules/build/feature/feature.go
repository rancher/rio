package feature

import (
	"context"

	"github.com/rancher/rio/modules/build/controllers/build"
	"github.com/rancher/rio/modules/build/controllers/gitcommit"
	"github.com/rancher/rio/modules/build/controllers/proxy"
	"github.com/rancher/rio/modules/build/controllers/service"
	stack1 "github.com/rancher/rio/modules/build/controllers/stack"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Register(ctx context.Context, rContext *types.Context) error {
	persistent := "false"
	scs, err := rContext.Storage.Storage().V1().StorageClass().List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, sc := range scs.Items {
		if sc.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
			persistent = "true"
			break
		}
	}

	apply := rContext.Apply.WithCacheTypes(rContext.Rio.Rio().V1().Service(), rContext.Core.Core().V1().ConfigMap())
	feature := &features.FeatureController{
		FeatureName: "build",
		FeatureSpec: features.FeatureSpec{
			Description: "Rio Build, from source code to deployment",
		},
		SystemStacks: []*stack.SystemStack{
			stack.NewSystemStack(apply, rContext.Admin.Admin().V1().SystemStack(), rContext.Namespace, "build"),
			stack.NewSystemStack(rContext.Apply, rContext.Admin.Admin().V1().SystemStack(), rContext.Namespace, "tekton"),
		},
		Controllers: []features.ControllerRegister{
			service.Register,
			build.Register,
			gitcommit.Register,
			proxy.Register,
			service.Register,
			stack1.Register,
		},
		FixedAnswers: map[string]string{
			"NAMESPACE":  rContext.Namespace,
			"PERSISTENT": persistent,
		},
	}
	return feature.Register()
}
