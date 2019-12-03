package server

import (
	"context"
	"strings"

	"github.com/rancher/rio/modules"
	"github.com/rancher/rio/pkg/config"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/pkg/controllers"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/pkg/webhook"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/crd"
	"github.com/rancher/wrangler/pkg/kubeconfig"
	"github.com/rancher/wrangler/pkg/leader"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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

	if err := bootstrapResources(rioContext, systemNamespace); err != nil {
		return err
	}

	if err := configureFeature(rioContext, systemNamespace); err != nil {
		return err
	}

	// setting up auth webhook
	w := webhook.New(rioContext, kubeConfig)
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

func bootstrapResources(rioContext *types.Context, systemNamespace string) error {
	controllerStack := stack.NewSystemStack(rioContext.Apply, rioContext.Admin.Admin().V1().SystemStack(), systemNamespace, "rio-bootstrap")
	answer := map[string]string{
		"NAMESPACE": systemNamespace,
	}
	return controllerStack.Deploy(answer)
}

func configureFeature(rioContext *types.Context, systemNamespace string) error {
	conf, err := config.GetConfig(systemNamespace, rioContext.Core.Core().V1().ConfigMap())
	if err != nil {
		return err
	}

	if conf.Features == nil {
		conf.Features = map[string]config.FeatureConfig{}
	}
	t := true
	if config.ConfigController.Features == "*" {
		f := conf.Features["*"]
		f.Enabled = &t
		conf.Features["*"] = f
	} else {
		for _, feature := range strings.Split(config.ConfigController.Features, ",") {
			if feature == "" {
				continue
			}
			f := conf.Features[feature]
			f.Enabled = &t
			conf.Features[feature] = f
		}
	}

	featureConfig, err := config.SetConfig(constructors.NewConfigMap(systemNamespace, config.ConfigName, v1.ConfigMap{}), conf)
	if err != nil {
		return err
	}

	if _, err := rioContext.Core.Core().V1().ConfigMap().Update(featureConfig); err != nil {
		if errors.IsNotFound(err) {
			if _, err = rioContext.Core.Core().V1().ConfigMap().Create(featureConfig); err != nil {
				return err
			}
		}
		return err
	}
	return err
}

func Types(ctx context.Context, config *rest.Config) error {
	factory, err := crd.NewFactoryFromClient(config)
	if err != nil {
		return err
	}

	factory.BatchCreateCRDs(ctx, getCRDs()...)

	return factory.BatchWait()
}
