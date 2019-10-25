package util

import (
	"errors"

	"github.com/rancher/rio/cli/pkg/clicontext"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/services"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListAppServicesFromServiceName(ctx *clicontext.CLIContext, serviceName string) ([]riov1.Service, error) {
	namespace := ctx.GetSetNamespace()
	service, err := ctx.ByID(serviceName)
	if err != nil {
		return []riov1.Service{}, err
	}
	svc := service.Object.(*riov1.Service)
	app, _ := services.AppAndVersion(svc)
	if app == "" {
		return []riov1.Service{}, errors.New("invalid app for service")
	}
	svcs, err := ctx.Rio.Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		return []riov1.Service{}, err
	}
	var revisions []riov1.Service
	for _, rev := range svcs.Items {
		if app == rev.Spec.App || app == rev.Name {
			revisions = append(revisions, rev)
		}
	}
	if len(revisions) == 0 {
		return []riov1.Service{}, errors.New("no services found")
	}
	return revisions, nil
}
