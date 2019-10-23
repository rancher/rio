package util

import (
	"fmt"

	"github.com/rancher/rio/cli/pkg/clicontext"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListPods(ctx *clicontext.CLIContext, services ...string) ([]corev1.Pod, error) {
	var pods []corev1.Pod

	for _, service := range services {
		pds, err := ctx.Core.Pods(ctx.GetSetNamespace()).List(metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", "rio.cattle.io/service", service),
		})
		if err != nil {
			return nil, err
		}

		pods = append(pods, pds.Items...)
	}
	return pods, nil
}
