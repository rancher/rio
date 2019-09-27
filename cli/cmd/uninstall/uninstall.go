package uninstall

import (
	"fmt"
	"strings"
	"time"

	certmanagerv1alpha1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	buildv1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	gitwatcherv1 "github.com/rancher/gitwatcher/pkg/apis/gitwatcher.cattle.io/v1"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/up/questions"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Uninstall struct {
	Namespace string `desc:"namespace to install system resources" default:"rio-system"`
}

func (u Uninstall) Run(ctx *clicontext.CLIContext) error {
	if ctx.K8s == nil {
		return fmt.Errorf("can't contact Kubernetes cluster. Please make sure your cluster is accessible")
	}

	_, err := ctx.Core.ConfigMaps("linkerd").Get("linkerd-config", metav1.GetOptions{})
	if err == nil {
		if err := u.uninstallLinkerd(ctx); err != nil {
			return err
		}
	}

	var systemNamespace string
	rioInfo, err := ctx.Project.RioInfos().Get("rio", metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if rioInfo.Status.SystemNamespace != "" {
		systemNamespace = rioInfo.Status.SystemNamespace
	} else {
		systemNamespace = u.Namespace
	}

	fmt.Println("Deleting Rio management controller...")
	if err := ctx.K8s.AppsV1().Deployments(systemNamespace).Delete("rio-controller", &metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
		return err
	}
	fmt.Println("Waiting for Rio controller pod to be deleted...")
	for {
		pods, err := ctx.Core.Pods(systemNamespace).List(metav1.ListOptions{
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

	fmt.Println("Deleting system component services")
	for systemSvc := range rioInfo.Status.SystemComponentReadyMap {
		if err := ctx.K8s.AppsV1().Deployments(systemNamespace).Delete(systemSvc, &metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
			return err
		}
	}

	fmt.Println("Cleaning up finalizers for resource PublicDomain, group admin.rio.cattle.io...")
	pds, err := ctx.Project.PublicDomains().List(metav1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	for _, pd := range pds.Items {
		pd.Finalizers = nil
		if _, err := ctx.Project.PublicDomains().Update(&pd); err != nil && !errors.IsNotFound(err) {
			return err
		}
	}

	// rio.cattle.io
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

	fmt.Println("Cleaning up finalizers for resource Build, group tekton.dev...")
	builds, err := ctx.Build.TaskRuns("").List(metav1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	for _, build := range builds.Items {
		build.Finalizers = nil
		if _, err := ctx.Build.TaskRuns(build.Namespace).Update(&build); err != nil && !errors.IsNotFound(err) {
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

	if ok, err := confirmDelete("config.istio.io/v1alpha2"); err != nil {
		return err
	} else if ok {
		configv1resources, err := ctx.K8s.Discovery().ServerResourcesForGroupVersion("config.istio.io/v1alpha2")
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
		if configv1resources != nil {
			for _, resource := range configv1resources.APIResources {
				toDelete = append(toDelete, fmt.Sprintf("%s.%s", resource.Name, "config.istio.io"))
			}
		}
	}

	if ok, err := confirmDelete(buildv1alpha1.SchemeGroupVersion.String()); err != nil {
		return err
	} else if ok {
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
	}

	if ok, err := confirmDelete("caching.internal.knative.dev/v1alpha1"); err != nil {
		return err
	} else if ok {
		knativeinternalresources, err := ctx.K8s.Discovery().ServerResourcesForGroupVersion("caching.internal.knative.dev/v1alpha1")
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
		if knativeinternalresources != nil {
			for _, resource := range knativeinternalresources.APIResources {
				toDelete = append(toDelete, fmt.Sprintf("%s.%s", resource.Name, "caching.internal.knative.dev"))
			}
		}
	}

	if ok, err := confirmDelete(certmanagerv1alpha1.SchemeGroupVersion.String()); err != nil {
		return err
	} else if ok {
		certmanagerresource, err := ctx.K8s.Discovery().ServerResourcesForGroupVersion(certmanagerv1alpha1.SchemeGroupVersion.String())
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
		if certmanagerresource != nil {
			for _, resource := range certmanagerresource.APIResources {
				toDelete = append(toDelete, fmt.Sprintf("%s.%s", resource.Name, certmanagerv1alpha1.SchemeGroupVersion.Group))
			}
		}
	}

	if ok, err := confirmDelete("rbac.istio.io/v1alpha1"); err != nil {
		return err
	} else if ok {
		rbacistioresource, err := ctx.K8s.Discovery().ServerResourcesForGroupVersion("rbac.istio.io/v1alpha1")
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
		if rbacistioresource != nil {
			for _, resource := range rbacistioresource.APIResources {
				toDelete = append(toDelete, fmt.Sprintf("%s.%s", resource.Name, "rbac.istio.io"))
			}
		}
	}

	if ok, err := confirmDelete("authentication.istio.io/v1alpha1"); err != nil {
		return err
	} else if ok {
		authistioresource, err := ctx.K8s.Discovery().ServerResourcesForGroupVersion("authentication.istio.io/v1alpha1")
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
		if authistioresource != nil {
			for _, resource := range authistioresource.APIResources {
				toDelete = append(toDelete, fmt.Sprintf("%s.%s", resource.Name, "authentication.istio.io"))
			}
		}
	}

	if ok, err := confirmDelete("networking.istio.io/v1alpha3"); err != nil {
		return err
	} else if ok {
		networkingistioresource, err := ctx.K8s.Discovery().ServerResourcesForGroupVersion("networking.istio.io/v1alpha3")
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
		if networkingistioresource != nil {
			for _, resource := range networkingistioresource.APIResources {
				toDelete = append(toDelete, fmt.Sprintf("%s.%s", resource.Name, "networking.istio.io"))
			}
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

	for _, del := range toDelete {
		if strings.Contains(del, "/") {
			continue
		}
		if err := crdclient.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(del, &metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
			return err
		}
	}

	fmt.Printf("Deleting Namespace %s...\n", systemNamespace)
	if err := ctx.Core.Namespaces().Delete(systemNamespace, &metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
		return err
	}

	fmt.Println("Rio is uninstalled from your cluster")
	return nil
}

func confirmDelete(resource string) (bool, error) {
	msg := fmt.Sprintf("Do you want to delete CRD group %v? some CRDs might be shared with your existing system components(istio, knative, cert-manager)\n", resource)

	options := []string{
		"[1] Yes\n",
		"[2] No\n",
	}

	num, err := questions.PromptOptions(msg, 1, options...)
	if err != nil {
		return false, err
	}

	if num == 0 {
		fmt.Printf("Adding CRD %v to delete list\n", resource)
		return true, nil
	}
	return false, err
}

func (u Uninstall) uninstallLinkerd(ctx *clicontext.CLIContext) error {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    u.Namespace,
			GenerateName: "linkerd-uninstall-",
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &[]int32{120}[0],
			BackoffLimit:            &[]int32{1}[0],
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					ServiceAccountName: "rio-controller-serviceaccount",
					RestartPolicy:      v1.RestartPolicyNever,
					Containers: []v1.Container{
						{
							Name:            "linkerd-install",
							Image:           constants.LinkerdInstallImage,
							ImagePullPolicy: v1.PullAlways,
							Env: []v1.EnvVar{
								{
									Name:  "LINKERD_UNINSTALL",
									Value: "TRUE",
								},
							},
						},
					},
				},
			},
		},
	}
	existingJob, err := ctx.K8s.BatchV1().Jobs(u.Namespace).Create(job)
	if err != nil {
		return err
	}
	startTime := time.Now()
	fmt.Println("Waiting for linkerd uninstall job to be finished")
	for {
		job, err := ctx.K8s.BatchV1().Jobs(u.Namespace).Get(existingJob.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if time.Now().After(startTime.Add(time.Minute * 2)) {
			return fmt.Errorf("Timeout waiting for linkerd uninstall job")
		}
		if job.Status.CompletionTime == nil {
			time.Sleep(time.Second * 5)
			continue
		}
		break
	}
	return nil
}
