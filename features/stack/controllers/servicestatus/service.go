package servicestatus

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	"github.com/rancher/norman/condition"
	"github.com/rancher/rio/pkg/serviceset"
	"github.com/rancher/rio/types"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	appsv1 "k8s.io/api/apps/v1beta2"
	"k8s.io/apimachinery/pkg/labels"
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
		serviceLister: rContext.Rio.Service.Cache(),
		services:      rContext.Rio.Service,
	}

	rContext.Apps.Deployment.OnChange(ctx, "sub-service-deploy-controller", s.deploymentChanged)
	rContext.Apps.DaemonSet.OnChange(ctx, "sub-service-daemonset-controller", s.daemonSetChanged)
	rContext.Apps.StatefulSet.OnChange(ctx, "sub-service-sts-controller", s.statefulSetChanged)

	rContext.Rio.Service.OnChange(ctx, "service-promote-controller", s.promote)

	return nil
}

type subServiceController struct {
	services      riov1.ServiceClient
	serviceLister riov1.ServiceClientCache
}

func (s *subServiceController) getService(labels map[string]string) *riov1.Service {
	name := labels["rio.cattle.io/service-name"]
	stackName := labels["rio.cattle.io/stack"]

	svc, err := s.serviceLister.Get(stackName, name)
	if err != nil {
		return nil
	}
	return svc
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
	} else if hasAvailableDS(newService.Status.DaemonSetStatus) {
		newService.Status.Conditions = nil
	} else if hasAvailableSS(newService.Status.StatefulSetStatus) {
		newService.Status.Conditions = nil
		if newService.Status.StatefulSetStatus == nil || len(newService.Status.StatefulSetStatus.Conditions) == 0 {
			riov1.PendingCondition.True(newService)
		}
	}

	if !reflect.DeepEqual(service.Status, newService.Status) {
		_, err := s.services.Update(newService)
		return err
	}

	return nil
}

func (s *subServiceController) daemonSetChanged(dep *appsv1.DaemonSet) (runtime.Object, error) {
	service := s.getService(dep.Labels)
	if service == nil {
		return nil, nil
	}

	newService := service.DeepCopy()
	newService.Status.DaemonSetStatus = dep.Status.DeepCopy()
	newService.Status.DaemonSetStatus.ObservedGeneration = 0

	return nil, s.updateStatus(service, newService, dep, dep.Generation, dep.Status.ObservedGeneration)
}

func (s *subServiceController) statefulSetChanged(dep *appsv1.StatefulSet) (runtime.Object, error) {
	service := s.getService(dep.Labels)
	if service == nil {
		return nil, nil
	}

	newService := service.DeepCopy()
	newService.Status.StatefulSetStatus = dep.Status.DeepCopy()
	newService.Status.StatefulSetStatus.ObservedGeneration = 0

	return nil, s.updateStatus(service, newService, dep, dep.Generation, dep.Status.ObservedGeneration)
}

func (s *subServiceController) deploymentChanged(dep *appsv1.Deployment) (runtime.Object, error) {
	service := s.getService(dep.Labels)
	if service == nil {
		return nil, nil
	}

	newService := service.DeepCopy()
	newService.Status.DeploymentStatus = dep.Status.DeepCopy()
	newService.Status.DeploymentStatus.ObservedGeneration = 0

	return nil, s.updateStatus(service, newService, dep, dep.Generation, dep.Status.ObservedGeneration)
}

func (s *subServiceController) promote(service *riov1.Service) (runtime.Object, error) {
	if service.Spec.Revision.ParentService == "" || !service.Spec.Revision.Promote {
		return nil, nil
	}

	services, err := s.serviceLister.List(service.Namespace, labels.Everything())
	if err != nil {
		return nil, err
	}

	serviceSets, err := serviceset.CollectionServices(services)
	if err != nil {
		return nil, err
	}

	serviceSet, ok := serviceSets[service.Spec.Revision.ParentService]
	if !ok {
		return nil, err
	}

	base := serviceSet.Service
	if base == nil {
		return nil, nil
	}

	for _, rev := range serviceSet.Revisions {
		if rev.Spec.Revision.Promote {
			newRev := rev.DeepCopy()
			newRev.Name = base.Name
			newRev.UID = base.UID
			newRev.ResourceVersion = base.ResourceVersion
			newRev.Spec.Revision.ParentService = ""
			newRev.Spec.Revision.Promote = false
			newRev.Spec.Revision.Weight = 0
			if _, err := s.services.Update(newRev); err != nil {
				return nil, errors.Wrapf(err, "failed to promote %s/%s/", rev.Namespace, rev.Name)
			}
			return nil, s.services.Delete(service.Namespace, service.Name, nil)
		}
	}

	return nil, nil
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

func hasAvailableDS(status *appsv1.DaemonSetStatus) bool {
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

func hasAvailableSS(status *appsv1.StatefulSetStatus) bool {
	if status != nil {
		cond := status.Conditions
		for _, c := range cond {
			if c.Type == "Available" {
				return true
			}
		}
	}
	if status.Replicas == status.ReadyReplicas {
		return true
	}
	return false
}
