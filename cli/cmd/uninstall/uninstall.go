package uninstall

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"

	certmanagerv1alpha1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	buildv1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	gitwatcherv1 "github.com/rancher/gitwatcher/pkg/apis/gitwatcher.cattle.io/v1"
	"github.com/rancher/rio/cli/pkg/clicontext"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	autoscalev1 "github.com/rancher/rio/pkg/apis/autoscale.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Uninstall struct {
	Namespace string `desc:"namespace to install system resources" default:"rio-system"`
}

func (u Uninstall) Run(ctx *clicontext.CLIContext) error {
	if ctx.K8s == nil {
		return fmt.Errorf("can't contact Kubernetes cluster. Please make sure your cluster is accessable")
	}

	fmt.Println("Deleting Rio management controller...")
	if err := ctx.K8s.AppsV1().Deployments(u.Namespace).Delete("rio-controller", &metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
		return err
	}
	fmt.Println("Waiting for Rio controller pod to be deleted...")
	for {
		pods, err := ctx.Core.Pods(u.Namespace).List(metav1.ListOptions{
			LabelSelector: "rio-controller=true",
		})
		if err != nil {
			return err
		}
		if len(pods.Items) > 0 {
			time.Sleep(time.Second * 2)
			continue
		}
		break
	}

	// admin.rio.cattle.io
	fmt.Println("Cleaning up finalizers for resource Feature, group admin.rio.cattle.io...")
	features, err := ctx.Project.Features("").List(metav1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	for _, f := range features.Items {
		f.Finalizers = nil
		if _, err := ctx.Project.Features(u.Namespace).Update(&f); err != nil && !errors.IsNotFound(err) {
			return err
		}
	}
	fmt.Println("Cleaning up finalizers for resource PublicDomain, group admin.rio.cattle.io...")
	pds, err := ctx.Project.PublicDomains("").List(metav1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	for _, pd := range pds.Items {
		pd.Finalizers = nil
		if _, err := ctx.Project.PublicDomains(pd.Namespace).Update(&pd); err != nil && !errors.IsNotFound(err) {
			return err
		}
	}

	// rio.cattle.io
	fmt.Println("Cleaning up finalizers for resource App, group rio.cattle.io...")
	apps, err := ctx.Rio.Apps("").List(metav1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	for _, app := range apps.Items {
		app.Finalizers = nil
		if _, err := ctx.Rio.Apps(app.Namespace).Update(&app); err != nil && !errors.IsNotFound(err) {
			return err
		}
	}

	fmt.Println("Cleaning up finalizers for resource Service, group rio.cattle.io...")
	services, err := ctx.Rio.Services("").List(metav1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	for _, service := range services.Items {
		service.Finalizers = nil
		if _, err := ctx.Rio.Services(service.Namespace).Update(&service); err != nil && !errors.IsNotFound(err) {
			return err
		}
	}

	fmt.Println("Cleaning up finalizers for resource Router, group rio.cattle.io...")
	routers, err := ctx.Rio.Routers("").List(metav1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	for _, router := range routers.Items {
		router.Finalizers = nil
		if _, err := ctx.Rio.Routers(router.Namespace).Update(&router); err != nil && !errors.IsNotFound(err) {
			return err
		}
	}

	fmt.Println("Cleaning up finalizers for resource ExternalService, group rio.cattle.io...")
	ess, err := ctx.Rio.ExternalServices("").List(metav1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	for _, es := range ess.Items {
		es.Finalizers = nil
		if _, err := ctx.Rio.ExternalServices(es.Namespace).Update(&es); err != nil && !errors.IsNotFound(err) {
			return err
		}
	}

	fmt.Println("Cleaning up finalizers for resource Build, group build.knative.dev...")
	builds, err := ctx.Build.Builds("").List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, build := range builds.Items {
		build.Finalizers = nil
		if _, err := ctx.Build.Builds(build.Namespace).Update(&build); err != nil && !errors.IsNotFound(err) {
			return err
		}
	}

	fmt.Println("Cleaning up finalizers for resource ServiceRecommendations, group autoscale.rio.cattle.io...")
	autoscales, err := ctx.Autoscale.ServiceScaleRecommendations("").List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, ssr := range autoscales.Items {
		ssr.Finalizers = nil
		if _, err := ctx.Autoscale.ServiceScaleRecommendations(ssr.Namespace).Update(&ssr); err != nil && !errors.IsNotFound(err) {
			return err
		}
	}

	crdclient, err := clientset.NewForConfig(ctx.RestConfig)
	if err != nil {
		return err
	}
	fmt.Println("Cleaning up CRDs...")
	var toDelete []string
	riov1resources, err := ctx.K8s.Discovery().ServerResourcesForGroupVersion(riov1.SchemeGroupVersion.String())
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if riov1resources != nil {
		for _, resource := range riov1resources.APIResources {
			toDelete = append(toDelete, fmt.Sprintf("%s.%s", resource.Name, riov1.SchemeGroupVersion.Group))
		}
	}

	adminv1resources, err := ctx.K8s.Discovery().ServerResourcesForGroupVersion(adminv1.SchemeGroupVersion.String())
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if adminv1resources != nil {
		for _, resource := range adminv1resources.APIResources {
			toDelete = append(toDelete, fmt.Sprintf("%s.%s", resource.Name, adminv1.SchemeGroupVersion.Group))
		}
	}

	configv1resources, err := ctx.K8s.Discovery().ServerResourcesForGroupVersion("config.istio.io/v1alpha2")
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if configv1resources != nil {
		for _, resource := range configv1resources.APIResources {
			toDelete = append(toDelete, fmt.Sprintf("%s.%s", resource.Name, "config.istio.io"))
		}
	}

	knativebuildresource, err := ctx.K8s.Discovery().ServerResourcesForGroupVersion(buildv1alpha1.SchemeGroupVersion.String())
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if knativebuildresource != nil {
		for _, resource := range knativebuildresource.APIResources {
			if strings.Contains(resource.Name, "/") {
				continue
			}
			toDelete = append(toDelete, fmt.Sprintf("%s.%s", resource.Name, buildv1alpha1.SchemeGroupVersion.Group))
		}
	}

	knativeinternalresources, err := ctx.K8s.Discovery().ServerResourcesForGroupVersion("caching.internal.knative.dev/v1alpha1")
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if knativeinternalresources != nil {
		for _, resource := range knativeinternalresources.APIResources {
			toDelete = append(toDelete, fmt.Sprintf("%s.%s", resource.Name, "caching.internal.knative.dev"))
		}
	}

	certmanagerresource, err := ctx.K8s.Discovery().ServerResourcesForGroupVersion(certmanagerv1alpha1.SchemeGroupVersion.String())
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if certmanagerresource != nil {
		for _, resource := range certmanagerresource.APIResources {
			toDelete = append(toDelete, fmt.Sprintf("%s.%s", resource.Name, certmanagerv1alpha1.SchemeGroupVersion.Group))
		}
	}

	rbacistioresource, err := ctx.K8s.Discovery().ServerResourcesForGroupVersion("rbac.istio.io/v1alpha1")
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if rbacistioresource != nil {
		for _, resource := range rbacistioresource.APIResources {
			toDelete = append(toDelete, fmt.Sprintf("%s.%s", resource.Name, "rbac.istio.io"))
		}
	}

	authistioresource, err := ctx.K8s.Discovery().ServerResourcesForGroupVersion("authentication.istio.io/v1alpha1")
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if authistioresource != nil {
		for _, resource := range authistioresource.APIResources {
			toDelete = append(toDelete, fmt.Sprintf("%s.%s", resource.Name, "authentication.istio.io"))
		}
	}

	networkingistioresource, err := ctx.K8s.Discovery().ServerResourcesForGroupVersion("networking.istio.io/v1alpha3")
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if networkingistioresource != nil {
		for _, resource := range networkingistioresource.APIResources {
			toDelete = append(toDelete, fmt.Sprintf("%s.%s", resource.Name, "networking.istio.io"))
		}
	}

	gitwatcherresources, err := ctx.K8s.Discovery().ServerResourcesForGroupVersion(gitwatcherv1.SchemeGroupVersion.String())
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if gitwatcherresources != nil {
		for _, resource := range gitwatcherresources.APIResources {
			toDelete = append(toDelete, fmt.Sprintf("%s.%s", resource.Name, gitwatcherv1.SchemeGroupVersion.Group))
		}
	}

	autoscalerresources, err := ctx.K8s.Discovery().ServerResourcesForGroupVersion(autoscalev1.SchemeGroupVersion.String())
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if autoscalerresources != nil {
		for _, resource := range autoscalerresources.APIResources {
			toDelete = append(toDelete, fmt.Sprintf("%s.%s", resource.Name, autoscalev1.SchemeGroupVersion.Group))
		}
	}

	for _, del := range toDelete {
		if err := crdclient.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(del, &metav1.DeleteOptions{}); err != nil {
			return err
		}
	}

	fmt.Printf("Deleting Namespace %s...\n", u.Namespace)
	if err := ctx.Core.Namespaces().Delete(u.Namespace, &metav1.DeleteOptions{}); err != nil {
		return err
	}

	fmt.Println("Rio is uninstalled from your cluster")
	return nil
}
