package util

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/services"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListAppServicesFromAppName(ctx *clicontext.CLIContext, namespace, appName string) ([]*riov1.Service, error) {
	svcs, err := ctx.Rio.Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		return []*riov1.Service{}, err
	}
	var revisions []*riov1.Service
	for i := range svcs.Items {
		rev := svcs.Items[i]
		revApp, _ := services.AppAndVersion(&rev)
		if revApp == appName {
			revisions = append(revisions, &rev)
		}
	}
	return revisions, nil
}
