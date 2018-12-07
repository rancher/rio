package service

import (
	"context"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/norman/condition"
	"github.com/rancher/rio/pkg/apply"
	"github.com/rancher/rio/pkg/deploy/istio/output"
	"github.com/rancher/rio/pkg/deploy/stack/populate/istio"
	service2 "github.com/rancher/rio/pkg/deploy/stack/populate/service"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"github.com/sirupsen/logrus"
	"istio.io/api/networking/v1alpha3"
	"istio.io/istio/pilot/pkg/model"
	appsv1 "k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	rContext.Apps.Deployment.OnChange(ctx, "sub-service-controller", s.deploymentChanged)
	rContext.Apps.DaemonSet.OnChange(ctx, "sub-service-controller", s.daemonSetChanged)
	rContext.Apps.StatefulSet.OnChange(ctx, "sub-service-controller", s.statefulSetChanged)

	rContext.Rio.Service.OnChange(ctx, "service-controller", s.promote)

	return nil
}

type subServiceController struct {
	services      riov1.ServiceClient
	serviceLister riov1.ServiceClientCache
}

func (s *subServiceController) getService(ns string, labels map[string]string) *riov1.Service {
	name := labels["rio.cattle.io/service-name"]

	svc, err := s.serviceLister.Get(ns, name)
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
	}

	if !reflect.DeepEqual(service.Status, newService.Status) {
		_, err := s.services.Update(newService)
		return err
	}

	return nil
}

func (s *subServiceController) daemonSetChanged(dep *appsv1.DaemonSet) (runtime.Object, error) {
	service := s.getService(dep.Namespace, dep.Labels)
	if service == nil {
		return nil, nil
	}

	newService := service.DeepCopy()
	newService.Status.DaemonSetStatus = &dep.Status

	return nil, s.updateStatus(service, newService, dep, dep.Generation, dep.Status.ObservedGeneration)
}

func (s *subServiceController) statefulSetChanged(dep *appsv1.StatefulSet) (runtime.Object, error) {
	service := s.getService(dep.Namespace, dep.Labels)
	if service == nil {
		return nil, nil
	}

	newService := service.DeepCopy()
	newService.Status.StatefulSetStatus = &dep.Status

	return nil, s.updateStatus(service, newService, dep, dep.Generation, dep.Status.ObservedGeneration)
}

func (s *subServiceController) deploymentChanged(dep *appsv1.Deployment) (runtime.Object, error) {
	service := s.getService(dep.Namespace, dep.Labels)
	if service == nil {
		return nil, nil
	}

	newService := service.DeepCopy()
	newService.Status.DeploymentStatus = &dep.Status

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

	serviceSets, err := service2.CollectionServices(services)
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
	return false
}

func (s subServiceController) acmeSolver(key string, service *v1.Service) error {
	if service == nil {
		return nil
	}
	if strings.HasPrefix(service.Name, "cm-acme-http-solver-") {
		vs := acmeVirtualService(service.Annotations["acme.domain"], service.Name)
		ds := acmeDestinationRule(service.Name, service.Labels)
		return apply.Apply([]runtime.Object{vs, ds}, nil, settings.RioSystemNamespace, key)
	}
	return nil
}

func acmeVirtualService(domain, host string) runtime.Object {
	vss := &v1alpha3.VirtualService{
		Gateways: []string{istio.GetPublicGateway()},
		Hosts:    []string{domain},
	}
	httpMatch := &v1alpha3.HTTPMatchRequest{}
	httpMatch.Uri = &v1alpha3.StringMatch{
		MatchType: &v1alpha3.StringMatch_Prefix{
			Prefix: "/.well-known/acme-challenge/",
		},
	}
	httpRoute := &v1alpha3.HTTPRoute{}
	httpRoute.Route = []*v1alpha3.DestinationWeight{
		{
			Destination: &v1alpha3.Destination{
				Host: host,
				Port: &v1alpha3.PortSelector{
					Port: &v1alpha3.PortSelector_Number{
						Number: uint32(8089),
					},
				},
				Subset: host,
			},
		},
	}
	httpRoute.Match = []*v1alpha3.HTTPMatchRequest{
		httpMatch,
	}
	vss.Http = []*v1alpha3.HTTPRoute{
		httpRoute,
	}
	vs := &output.VirtualService{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "VirtualService",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      host,
			Namespace: settings.RioSystemNamespace,
		},
	}
	m, err := model.ToJSONMap(vss)
	if err != nil {
		logrus.Error(err)
	}
	vs.Spec = m
	return vs
}

func acmeDestinationRule(host string, labels map[string]string) runtime.Object {
	des := &output.DestinationRule{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "DestinationRule",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      host,
			Namespace: settings.RioSystemNamespace,
		},
	}
	desspec := &v1alpha3.DestinationRule{
		Host: host,
		Subsets: []*v1alpha3.Subset{
			{
				Labels: labels,
				Name:   host,
			},
		},
	}
	m, err := model.ToJSONMap(desspec)
	if err != nil {
		logrus.Error(err)
	}
	des.Spec = m
	return des
}
