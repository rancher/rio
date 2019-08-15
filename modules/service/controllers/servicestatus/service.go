package servicestatus

import (
	"context"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/condition"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	progressing = condition.Cond("Progressing")
	updated     = condition.Cond("Updated")
	upgrading   = map[string]bool{
		"ReplicaSetUpdated":    true,
		"NewReplicaSetCreated": true,
	}
)

func Register(ctx context.Context, rContext *types.Context) error {
	s := &subServiceController{
		serviceLister: rContext.Rio.Rio().V1().Service().Cache(),
		services:      rContext.Rio.Rio().V1().Service(),
	}

	rContext.Apps.Apps().V1().Deployment().OnChange(ctx, "sub-service-deploy-controller", s.deploymentChanged)

	rContext.Apps.Apps().V1().DaemonSet().OnChange(ctx, "sub-service-deploy-daemonset-controller", s.daemonsetChanged)

	return nil
}

type subServiceController struct {
	services      v1.ServiceController
	serviceLister v1.ServiceCache
}

func (s *subServiceController) updateStatus(service, newService *riov1.Service, dep runtime.Object, generation, observedGeneration int64) error {
	isUpgrading := false

	if upgrading[progressing.GetReason(dep)] || generation != observedGeneration {
		isUpgrading = true
	}

	if isUpgrading {
		updated.Unknown(newService)
	} else if hasAvailable(newService.Status.DeploymentStatus) {
		newService.Status.Conditions = nil
	}

	if !equality.Semantic.DeepEqual(service.Status, newService.Status) {
		_, err := s.services.Update(newService)
		return err
	}

	return nil
}

func (s *subServiceController) daemonsetChanged(key string, ds *appsv1.DaemonSet) (*appsv1.DaemonSet, error) {
	if ds == nil {
		return ds, nil
	}
	if ds.DeletionTimestamp != nil {
		return ds, nil
	}
	service, err := s.serviceLister.Get(ds.Namespace, ds.Name)
	if errors.IsNotFound(err) {
		return ds, nil
	} else if err != nil {
		return ds, err
	}

	if service.DeletionTimestamp != nil {
		return ds, nil
	}

	newService := service.DeepCopy()
	newService.Status.ScaleStatus = &riov1.ScaleStatus{
		Ready:       int(ds.Status.NumberReady),
		Unavailable: int(ds.Status.NumberUnavailable),
		Available:   int(ds.Status.NumberUnavailable),
		Updated:     int(ds.Status.NumberReady),
	}

	_, err = s.services.Update(newService)
	return ds, err
}

func (s *subServiceController) deploymentChanged(key string, dep *appsv1.Deployment) (*appsv1.Deployment, error) {
	if dep == nil {
		return nil, nil
	}
	if dep.DeletionTimestamp != nil {
		return nil, nil
	}
	service, err := s.serviceLister.Get(dep.Namespace, dep.Name)
	if errors.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	if service.DeletionTimestamp != nil {
		return dep, nil
	}

	newService := service.DeepCopy()
	newService.Status.DeploymentStatus = dep.Status.DeepCopy()
	newService.Status.DeploymentStatus.ObservedGeneration = 0
	newService.Status.ScaleStatus = &riov1.ScaleStatus{
		Ready:       int(dep.Status.ReadyReplicas),
		Unavailable: int(dep.Status.UnavailableReplicas),
		Available:   int(dep.Status.AvailableReplicas - dep.Status.ReadyReplicas),
		Updated:     int(dep.Status.UpdatedReplicas),
	}

	return nil, s.updateStatus(service, newService, dep, dep.Generation, dep.Status.ObservedGeneration)
}

func hasAvailable(status *appsv1.DeploymentStatus) bool {
	if status != nil {
		cond := status.Conditions
		for _, c := range cond {
			if c.Type == "Available" {
				return true
			}
		}
	}
	return false
}
