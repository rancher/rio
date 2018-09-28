package service

import (
	"context"

	"reflect"

	"github.com/rancher/norman/condition"
	"github.com/rancher/rio/pkg/deploy"
	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"github.com/rancher/types/apis/apps/v1beta2"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	"k8s.io/apimachinery/pkg/labels"
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
	s := &serviceController{
		deploymentLister:  rContext.Apps.Deployments("").Controller().Lister(),
		statefulSetLister: rContext.Apps.StatefulSets("").Controller().Lister(),
		serviceController: rContext.Rio.Services("").Controller(),
		services:          rContext.Rio.Services(""),
	}

	s.serviceController.AddHandler("service-controller", s.sync)
	rContext.Apps.Deployments("").AddHandler("service-controller", s.deploymentChanged)
	rContext.Apps.StatefulSets("").AddHandler("service-controller", s.statefulSetChanged)
}

type serviceController struct {
	statefulSetLister v1beta2.StatefulSetLister
	deploymentLister  v1beta2.DeploymentLister
	serviceController v1beta1.ServiceController
	services          v1beta1.ServiceInterface
}

func (s *serviceController) deploymentChanged(key string, deployment *appsv1beta2.Deployment) error {
	if deployment == nil {
		return nil
	}

	service := deployment.Labels["rio.cattle.io/service"]
	ns := deployment.Labels["rio.cattle.io/namespace"]
	if ns != "" && service != "" {
		s.serviceController.Enqueue(ns, service)
	}

	return nil
}

func (s *serviceController) statefulSetChanged(key string, statefulSet *appsv1beta2.StatefulSet) error {
	if statefulSet == nil {
		return nil
	}

	service := statefulSet.Labels["rio.cattle.io/service"]
	ns := statefulSet.Labels["rio.cattle.io/namespace"]
	if ns != "" && service != "" {
		s.serviceController.Enqueue(ns, service)
	}

	return nil
}

func (s *serviceController) promote(service *v1beta1.Service) (bool, error) {
	for rev, revConfig := range service.Spec.Revisions {
		if !revConfig.Promote {
			continue
		}

		newService, err := deploy.MergeRevisionToService(service, rev)
		if err != nil {
			return false, err
		}

		service.Spec.ServiceUnversionedSpec = *newService
		service.Labels = newService.Labels
		delete(service.Spec.Revisions, rev)
		_, err = s.services.Update(service)
		return true, err
	}

	return false, nil
}

func (s *serviceController) sync(key string, service *v1beta1.Service) error {
	if service == nil {
		return nil
	}

	ok, err := s.promote(service)
	if err != nil || ok {
		return err
	}

	set := labels.Set{}
	set["rio.cattle.io/service"] = service.Name
	set["rio.cattle.io/namespace"] = service.Namespace

	isUpgrading := false
	newService := service.DeepCopy()

	deps, err := s.deploymentLister.List(service.Namespace, set.AsSelector())
	if err != nil {
		return err
	}
	if len(deps) != 0 {
		for _, dep := range deps {
			rev := dep.Labels["rio.cattle.io/revision"]
			if rev == "latest" {
				newService.Status.DeploymentStatus = &dep.Status
			} else {
				if revSpec, ok := newService.Spec.Revisions[rev]; ok {
					revSpec.Status.DeploymentStatus = &dep.Status
					newService.Spec.Revisions[rev] = revSpec
				}
			}

			if upgrading[progressing.GetReason(dep)] || dep.Generation != dep.Status.ObservedGeneration {
				isUpgrading = true
			}
		}

		if isUpgrading {
			updated.Unknown(newService)
		} else if hasAvailable(newService.Status.DeploymentStatus.Conditions) {
			newService.Status.Conditions = nil
		}
	} else {
		stss, err := s.statefulSetLister.List(service.Namespace, set.AsSelector())
		if err != nil {
			return err
		}
		for _, sts := range stss {
			rev := sts.Labels["rio.cattle.io/revision"]
			if rev == "latest" {
				newService.Status.DeploymentStatus = convertStatefulSetStatus(sts.Status)
			} else {
				if revSpec, ok := newService.Spec.Revisions[rev]; ok {
					revSpec.Status.DeploymentStatus = convertStatefulSetStatus(sts.Status)
					newService.Spec.Revisions[rev] = revSpec
				}
			}

			if upgrading[progressing.GetReason(sts)] || sts.Generation != sts.Status.ObservedGeneration {
				isUpgrading = true
			}
		}
		if isUpgrading {
			updated.Unknown(newService)
		} else {
			newService.Status.Conditions = nil
		}
	}

	if !reflect.DeepEqual(service, newService) {
		_, err := s.services.Update(newService)
		return err
	}

	return nil
}

func convertStatefulSetStatus(status appsv1beta2.StatefulSetStatus) *appsv1beta2.DeploymentStatus {
	conditions := make([]appsv1beta2.DeploymentCondition, 0)
	for _, c := range status.Conditions {
		conditions = append(conditions, convertStatefulSetCondition(c))
	}
	return &appsv1beta2.DeploymentStatus{
		Replicas:          status.Replicas,
		ReadyReplicas:     status.ReadyReplicas,
		UpdatedReplicas:   status.UpdatedReplicas,
		AvailableReplicas: status.CurrentReplicas,
		Conditions:        conditions,
	}
}

func convertStatefulSetCondition(condition appsv1beta2.StatefulSetCondition) appsv1beta2.DeploymentCondition {
	return appsv1beta2.DeploymentCondition{
		Type:               appsv1beta2.DeploymentConditionType(string(condition.Type)),
		Status:             condition.Status,
		LastTransitionTime: condition.LastTransitionTime,
		Reason:             condition.Reason,
		Message:            condition.Message,
	}
}

func hasAvailable(cond []appsv1beta2.DeploymentCondition) bool {
	for _, c := range cond {
		if c.Type == "Available" {
			return true
		}
	}
	return false
}
