package uninstall

import (
	"fmt"
	"strings"
	"time"

	gitwatcherv1 "github.com/rancher/gitwatcher/pkg/apis/gitwatcher.cattle.io/v1"
	"github.com/rancher/rio/cli/pkg/clicontext"
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

	gitwatcherresources, err := ctx.K8s.Discovery().ServerResourcesForGroupVersion(gitwatcherv1.SchemeGroupVersion.String())
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if gitwatcherresources != nil {
		for _, resource := range gitwatcherresources.APIResources {
			toDelete = append(toDelete, fmt.Sprintf("%s.%s", resource.Name, gitwatcherv1.SchemeGroupVersion.Group))
		}
	}

	crdclient, err := clientset.NewForConfig(ctx.RestConfig)
	if err != nil {
		return err
	}
	for _, del := range toDelete {
		if strings.Contains(del, "/") {
			continue
		}
		if err := crdclient.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(del, &metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
			return err
		}
	}

	fmt.Printf("Deleting System Namespace %s...\n", systemNamespace)
	if err := ctx.Core.Namespaces().Delete(systemNamespace, &metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
		return err
	}

	fmt.Println("Delete validating webhook")
	if err := ctx.K8s.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Delete(constants.AuthWebhookSecretName, &metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
		return err
	}

	fmt.Println("Rio is uninstalled from your cluster")
	return nil
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
