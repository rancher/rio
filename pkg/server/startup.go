package server

import (
	"context"

	"github.com/rancher/rio/modules"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/controllers"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/pkg/webhook"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/crd"
	"github.com/rancher/wrangler/pkg/kubeconfig"
	"github.com/rancher/wrangler/pkg/leader"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"
)

func Startup(ctx context.Context, systemNamespace, kubeConfig string) error {
	loader := kubeconfig.GetInteractiveClientConfig(kubeConfig)
	restConfig, err := loader.ClientConfig()
	if err != nil {
		return err
	}

	if err := Types(ctx, restConfig); err != nil {
		return err
	}

	ctx, rioContext := types.BuildContext(ctx, systemNamespace, restConfig)

	// detect and bootstrap developer environment
	devMode, err := bootstrapResources(rioContext, systemNamespace)
	if err != nil {
		return err
	}
	constants.DevMode = devMode

	// setting up auth webhook
	w := webhook.New(rioContext, kubeConfig, devMode)
	if err := w.Setup(); err != nil {
		return err
	}

	leader.RunOrDie(ctx, systemNamespace, "rio", rioContext.K8s, func(ctx context.Context) {
		runtime.Must(controllers.Register(ctx, rioContext))
		runtime.Must(modules.Register(ctx, rioContext))
		runtime.Must(rioContext.Start(ctx))
		<-ctx.Done()
	})

	return nil
}

func bootstrapResources(rioContext *types.Context, systemNamespace string) (bool, error) {
	if _, err := rioContext.Apps.Apps().V1().Deployment().Get(systemNamespace, "rio-controller", metav1.GetOptions{}); errors.IsNotFound(err) {
		controllerStack := stack.NewSystemStack(rioContext.Apply, systemNamespace, "rio-controller")
		answer := map[string]string{
			"NAMESPACE": systemNamespace,
		}
		if err := controllerStack.Deploy(answer); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func Types(ctx context.Context, config *rest.Config) error {
	factory, err := crd.NewFactoryFromClient(config)
	if err != nil {
		return err
	}

	factory.BatchCreateCRDs(ctx, getCRDs()...)

	return factory.BatchWait()
}
