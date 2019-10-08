package uninstall

import (
	"fmt"
	"time"

	certmanagerv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/up/questions"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
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

	if ok, err := confirmDelete(certmanagerv1alpha2.SchemeGroupVersion.String()); err != nil {
		return err
	} else if ok {
		certmanagerresource, err := ctx.K8s.Discovery().ServerResourcesForGroupVersion(certmanagerv1alpha2.SchemeGroupVersion.String())
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
		if certmanagerresource != nil {
			for _, resource := range certmanagerresource.APIResources {
				toDelete = append(toDelete, fmt.Sprintf("%s.%s", resource.Name, certmanagerv1alpha2.SchemeGroupVersion.Group))
			}
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
