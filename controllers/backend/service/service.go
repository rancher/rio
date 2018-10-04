package service

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	"github.com/rancher/norman/condition"
	service2 "github.com/rancher/rio/pkg/deploy/stack/populate/service"
	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
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

func Register(ctx context.Context, rContext *types.Context) {
	s := &subServiceController{
		serviceLister: rContext.Rio.Services("").Controller().Lister(),
		services:      rContext.Rio.Services(""),
	}

	rContext.Apps.Deployments("").AddHandler("sub-service-controller", s.deploymentChanged)
	rContext.Apps.DaemonSets("").AddHandler("sub-service-controller", s.daemonSetChanged)
	rContext.Apps.StatefulSets("").AddHandler("sub-service-controller", s.statefulSetChanged)

	rContext.Rio.Services("").AddHandler("service-controller", s.promote)
}

type subServiceController struct {
	services      v1beta1.ServiceInterface
	serviceLister v1beta1.ServiceLister
}

func (s *subServiceController) getService(ns string, labels map[string]string) *v1beta1.Service {
	name := labels["rio.cattle.io/service-name"]

	svc, err := s.serviceLister.Get(ns, name)
	if err != nil {
		return nil
	}
	return svc
}

func (s *subServiceController) updateStatus(service, newService *v1beta1.Service, dep runtime.Object, generation, observedGeneration int64) error {
	isUpgrading := false

	if upgrading[progressing.GetReason(dep)] || generation != observedGeneration {
		isUpgrading = true
	}

	if isUpgrading {
		updated.Unknown(newService)
	} else if hasAvailable(newService.Status.DeploymentStatus.Conditions) {
		newService.Status.Conditions = nil
	} else if hasAvailableDS(newService.Status.DaemonSetStatus.Conditions) {
		newService.Status.Conditions = nil
	} else if hasAvailableSS(newService.Status.StatefulSetStatus.Conditions) {
		newService.Status.Conditions = nil
	}

	if !reflect.DeepEqual(service.Status, newService.Status) {
		_, err := s.services.Update(newService)
		return err
	}

	return nil
}

func (s *subServiceController) daemonSetChanged(key string, dep *appsv1.DaemonSet) error {
	if dep == nil {
		return nil
	}

	service := s.getService(dep.Namespace, dep.Labels)
	if service == nil {
		return nil
	}

	newService := service.DeepCopy()
	newService.Status.DaemonSetStatus = &dep.Status

	return s.updateStatus(service, newService, dep, dep.Generation, dep.Status.ObservedGeneration)
}

func (s *subServiceController) statefulSetChanged(key string, dep *appsv1.StatefulSet) error {
	if dep == nil {
		return nil
	}

	service := s.getService(dep.Namespace, dep.Labels)
	if service == nil {
		return nil
	}

	newService := service.DeepCopy()
	newService.Status.StatefulSetStatus = &dep.Status

	return s.updateStatus(service, newService, dep, dep.Generation, dep.Status.ObservedGeneration)
}

func (s *subServiceController) deploymentChanged(key string, dep *appsv1.Deployment) error {
	if dep == nil {
		return nil
	}

	service := s.getService(dep.Namespace, dep.Labels)
	if service == nil {
		return nil
	}

	newService := service.DeepCopy()
	newService.Status.DeploymentStatus = &dep.Status

	return s.updateStatus(service, newService, dep, dep.Generation, dep.Status.ObservedGeneration)
}

func (s *subServiceController) promote(key string, service *v1beta1.Service) error {
	if service == nil {
		return nil
	}

	if service.Spec.Revision.ParentService == "" || !service.Spec.Revision.Promote {
		return nil
	}

	services, err := s.serviceLister.List(service.Namespace, labels.Everything())
	if err != nil {
		return err
	}

	serviceSets, err := service2.CollectionServices(services)
	if err != nil {
		return err
	}

	serviceSet, ok := serviceSets[service.Spec.Revision.ParentService]
	if !ok {
		return err
	}

	base := serviceSet.Service
	if base == nil {
		return nil
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
				return errors.Wrapf(err, "failed to promote %s/%s/", rev.Namespace, rev.Name)
			}
			return s.services.DeleteNamespaced(service.Namespace, service.Name, nil)
		}
	}

	return nil
}

func hasAvailable(cond []appsv1.DeploymentCondition) bool {
	for _, c := range cond {
		if c.Type == "Available" {
			return true
		}
	}
	return false
}

func hasAvailableDS(cond []appsv1.DaemonSetCondition) bool {
	for _, c := range cond {
		if c.Type == "Available" {
			return true
		}
	}
	return false
}

func hasAvailableSS(cond []appsv1.StatefulSetCondition) bool {
	for _, c := range cond {
		if c.Type == "Available" {
			return true
		}
	}
	return false
}
