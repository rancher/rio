package server

import (
	"context"

	"github.com/rancher/rio/modules"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/pkg/controllers"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/crd"
	"github.com/rancher/wrangler/pkg/kubeconfig"
	"github.com/rancher/wrangler/pkg/leader"
	v1 "k8s.io/api/core/v1"
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

	namespaceClient := rioContext.Core.Core().V1().Namespace()
	if _, err := namespaceClient.Get(systemNamespace, metav1.GetOptions{}); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		ns := constructors.NewNamespace(systemNamespace, v1.Namespace{})
		if _, err := namespaceClient.Create(ns); err != nil {
			return err
		}
	}

	leader.RunOrDie(ctx, systemNamespace, "rio", rioContext.K8s, func(ctx context.Context) {
		runtime.Must(controllers.Register(ctx, rioContext))
		runtime.Must(modules.Register(ctx, rioContext))
		runtime.Must(rioContext.Start(ctx))
		<-ctx.Done()
	})

	return nil
}

func Types(ctx context.Context, config *rest.Config) error {
	factory, err := crd.NewFactoryFromClient(config)
	if err != nil {
		return err
	}

	factory.BatchCreateCRDs(ctx, getCRDs()...)

	return factory.BatchWait()
}
