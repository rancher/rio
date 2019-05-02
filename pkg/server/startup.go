package server

import (
	"context"

	"github.com/rancher/rio/modules"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/pkg/controllers"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/crd"
	"github.com/rancher/wrangler/pkg/leader"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func Startup(ctx context.Context, systemNamespace, registry string, kubeConfig string) error {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return err
	}

	if err := Types(ctx, restConfig); err != nil {
		return err
	}

	ctx, rioContext := types.BuildContext(ctx, systemNamespace, registry, restConfig)

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

	crds := crd.NonNamespacedTypes(
		"Setting.project.rio.cattle.io/v1",
		"ClusterIssuer.certmanager.k8s.io/v1alpha1",
		"ClusterBuildTemplate.build.knative.dev/v1alpha1",
	)
	crds = append(crds, crd.NamespacedTypes(
		"BuildTemplate.build.knative.dev/v1alpha1",
		"Image.caching.internal.knative.dev/v1alpha1",

		"GitModule.git.rio.cattle.io/v1",

		"ExternalService.rio.cattle.io/v1",
		"Router.rio.cattle.io/v1",
		"Service.rio.cattle.io/v1",
		"PublicDomain.rio.cattle.io/v1",
		"App.rio.cattle.io/v1",

		"ClusterDomain.project.rio.cattle.io/v1",
		"Feature.project.rio.cattle.io/v1",
		"ListenConfig.project.rio.cattle.io/v1",

		"DestinationRule.networking.istio.io/v1alpha3",
		"Gateway.networking.istio.io/v1alpha3",
		"VirtualService.networking.istio.io/v1alpha3",
		"EgressRule.config.istio.io/v1alpha2",
		"RouteRule.config.istio.io/v1alpha2",
		"HTTPAPISpecBinding.config.istio.io/v1alpha2",
		"HTTPAPISpec.config.istio.io/v1alpha2",
		"QuotaSpecBinding.config.istio.io/v1alpha2",
		"QuotaSpec.config.istio.io/v1alpha2",
		"Policy.authentication.istio.io/v1alpha1",
		"metric.config.istio.io/v1alpha2",
		"rule.config.istio.io/v1alpha2",
		"prometheus.config.istio.io/v1alpha2",
		"kubernetes.config.istio.io/v1alpha2",
		"template.config.istio.io/v1alpha2",
		"attributemanifest.config.istio.io/v1alpha2",
		"instance.config.istio.io/v1alpha2",
		"adapter.config.istio.io/v1alpha2",
		"handler.config.istio.io/v1alpha2",
		"logentry.config.istio.io/v1alpha2",
		"stdio.config.istio.io/v1alpha2",
		"kubernetesenv.config.istio.io/v1alpha2",

		"GitWebHookExecution.webhookinator.rio.cattle.io/v1",
		"GitWebHookReceiver.webhookinator.rio.cattle.io/v1",
		"ServiceEntry.networking.istio.io/v1alpha3",

		"ServiceScaleRecommendation.autoscale.rio.cattle.io/v1",

		"Issuer.certmanager.k8s.io/v1alpha1",
		"Challenge.certmanager.k8s.io/v1alpha1",
		"Order.certmanager.k8s.io/v1alpha1",
		"Certificate.certmanager.k8s.io/v1alpha1",
	)...)

	factory.BatchCreateCRDs(ctx, crds...)

	return factory.BatchWait()
}
