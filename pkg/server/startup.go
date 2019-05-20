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

var Crds = append(crd.NonNamespacedTypes(
	"ClusterIssuer.certmanager.k8s.io/v1alpha1",

	"ClusterBuildTemplate.build.knative.dev/v1alpha1",

	"RioInfo.admin.rio.cattle.io/v1",
), crd.NamespacedTypes(
	"BuildTemplate.build.knative.dev/v1alpha1",
	"Image.caching.internal.knative.dev/v1alpha1",

	"App.rio.cattle.io/v1",
	"ExternalService.rio.cattle.io/v1",
	"Router.rio.cattle.io/v1",
	"Service.rio.cattle.io/v1",

	"ClusterDomain.admin.rio.cattle.io/v1",
	"Feature.admin.rio.cattle.io/v1",
	"ListenConfig.admin.rio.cattle.io/v1",
	"PublicDomain.admin.rio.cattle.io/v1",

	"DestinationRule.networking.istio.io/v1alpha3",
	"Gateway.networking.istio.io/v1alpha3",
	"ServiceEntry.networking.istio.io/v1alpha3",
	"VirtualService.networking.istio.io/v1alpha3",

	"adapter.config.istio.io/v1alpha2",
	"attributemanifest.config.istio.io/v1alpha2",
	"EgressRule.config.istio.io/v1alpha2",
	"handler.config.istio.io/v1alpha2",
	"HTTPAPISpecBinding.config.istio.io/v1alpha2",
	"HTTPAPISpec.config.istio.io/v1alpha2",
	"instance.config.istio.io/v1alpha2",
	"kubernetes.config.istio.io/v1alpha2",
	"kubernetesenv.config.istio.io/v1alpha2",
	"logentry.config.istio.io/v1alpha2",
	"metric.config.istio.io/v1alpha2",
	"Policy.authentication.istio.io/v1alpha1",
	"prometheus.config.istio.io/v1alpha2",
	"QuotaSpecBinding.config.istio.io/v1alpha2",
	"QuotaSpec.config.istio.io/v1alpha2",
	"RouteRule.config.istio.io/v1alpha2",
	"rule.config.istio.io/v1alpha2",
	"stdio.config.istio.io/v1alpha2",
	"template.config.istio.io/v1alpha2",

	"GitCommit.gitwatcher.cattle.io/v1",
	"GitWatcher.gitwatcher.cattle.io/v1",

	"ServiceScaleRecommendation.autoscale.rio.cattle.io/v1",

	"Certificate.certmanager.k8s.io/v1alpha1",
	"Challenge.certmanager.k8s.io/v1alpha1",
	"Issuer.certmanager.k8s.io/v1alpha1",
	"Order.certmanager.k8s.io/v1alpha1",
)...)

func Startup(ctx context.Context, systemNamespace, kubeConfig string) error {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
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

	factory.BatchCreateCRDs(ctx, Crds...)

	return factory.BatchWait()
}
