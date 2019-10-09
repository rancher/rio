package linkerd

import (
	"context"
	"fmt"
	"time"

	"github.com/rancher/rio/pkg/constants"

	"github.com/rancher/rio/modules/linkerd/feature"
	"github.com/rancher/rio/types"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Register(ctx context.Context, rContext *types.Context) error {
	if constants.ServiceMeshMode == constants.ServiceMeshModeLinkerd {
		if err := installLinkerd(rContext); err != nil {
			return err
		}
	}

	return feature.Register(ctx, rContext)
}

func installLinkerd(rContext *types.Context) error {
	cmClient := rContext.Core.Core().V1().ConfigMap()
	linkerdUpgrade := ""
	if _, err := cmClient.Get("linkerd", "linkerd-config", metav1.GetOptions{}); err == nil {
		linkerdUpgrade = "TRUE"
	}
	if constants.DevMode != "" && linkerdUpgrade == "TRUE" {
		return nil
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    rContext.Namespace,
			GenerateName: "linkerd-install-",
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
									Name:  "LINKERD_UPGRADE",
									Value: linkerdUpgrade,
								},
							},
						},
					},
				},
			},
		},
	}
	existingJob, err := rContext.Batch.Batch().V1().Job().Create(job)
	if err != nil {
		return err
	}

	startTime := time.Now()
	for {
		job, err := rContext.K8s.BatchV1().Jobs(rContext.Namespace).Get(existingJob.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if time.Now().After(startTime.Add(time.Minute * 10)) {
			return fmt.Errorf("timeout waiting for linkerd install job")
		}
		if job.Status.CompletionTime == nil {
			time.Sleep(time.Second * 20)
			continue
		}
		break
	}
	return nil
}
