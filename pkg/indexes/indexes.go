package indexes

import (
	"fmt"

	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/rio/types"
)

const (
	PublicDomainByTarget          = "byTarget"
	ClusterDomainByAssignedSecret = "bySecret"
	PublicDomainByAssignedSecret  = "bySecret"
	ServiceByApp                  = "apps"
)

func RegisterIndexes(rContext *types.Context) {
	publicDomain(rContext)
	secrets(rContext)
	service(rContext)
}

func publicDomain(rContext *types.Context) {
	rContext.Admin.Admin().V1().PublicDomain().Cache().AddIndexer(PublicDomainByTarget, func(obj *adminv1.PublicDomain) ([]string, error) {
		ns := obj.Spec.TargetNamespace
		if ns == "" {
			ns = obj.Namespace
		}

		var keys []string
		if obj.Spec.TargetApp == "" {
			return nil, nil
		}

		if obj.Spec.TargetVersion == "" {
			keys = append(keys, fmt.Sprintf("%s/%s", ns, obj.Spec.TargetApp))
		} else {
			keys = append(keys, fmt.Sprintf("%s/%s/%s", ns, obj.Spec.TargetApp, obj.Spec.TargetVersion))
		}

		return keys, nil
	})

	rContext.Admin.Admin().V1().PublicDomain().Cache().AddIndexer(PublicDomainByAssignedSecret, func(obj *adminv1.PublicDomain) ([]string, error) {
		if obj.Status.AssignedSecretName == "" {
			return nil, nil
		}
		return []string{
			fmt.Sprintf("%s/%s", obj.Namespace, obj.Status.AssignedSecretName),
		}, nil
	})
}

func secrets(rContext *types.Context) {
	rContext.Admin.Admin().V1().ClusterDomain().Cache().AddIndexer(ClusterDomainByAssignedSecret, func(obj *adminv1.ClusterDomain) ([]string, error) {
		if obj.Status.AssignedSecretName == "" {
			return nil, nil
		}
		return []string{
			obj.Status.AssignedSecretName,
		}, nil
	})
}

func service(rContext *types.Context) {
	rContext.Rio.Rio().V1().Service().Cache().AddIndexer(ServiceByApp, func(obj *riov1.Service) ([]string, error) {
		app, _ := services.AppAndVersion(obj)
		return []string{
			fmt.Sprintf("%s/%s", obj.Namespace, app),
		}, nil
	})
}
