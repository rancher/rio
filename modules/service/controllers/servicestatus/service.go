package servicestatus

import (
	"context"

	"github.com/rancher/rio/pkg/deployment"

	"github.com/rancher/rio/modules/service/pkg/endpoints"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	appsv1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/apps/v1"
	"github.com/rancher/wrangler/pkg/relatedresource"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	s := &statusHandler{
		daemonSetCache:   rContext.Apps.Apps().V1().DaemonSet().Cache(),
		deploymentCache:  rContext.Apps.Apps().V1().Deployment().Cache(),
		statefulSetCache: rContext.Apps.Apps().V1().StatefulSet().Cache(),
		resolver: endpoints.NewResolver(ctx, rContext.Namespace,
			rContext.Rio.Rio().V1().Service(),
			rContext.Rio.Rio().V1().Service().Cache(),
			rContext.Admin.Admin().V1().ClusterDomain(),
			rContext.Admin.Admin().V1().PublicDomain(),
		),
	}

	relatedresource.Watch(ctx, "service-status", findService,
		rContext.Rio.Rio().V1().Service(),
		rContext.Apps.Apps().V1().Deployment(),
		rContext.Apps.Apps().V1().DaemonSet(),
		rContext.Apps.Apps().V1().StatefulSet())

	v1.RegisterServiceStatusHandler(ctx,
		rContext.Rio.Rio().V1().Service(),
		"",
		"service",
		s.handle)

	return nil
}

func findService(namespace, name string, obj runtime.Object) ([]relatedresource.Key, error) {
	if obj == nil {
		return nil, nil
	}
	meta, err := meta.Accessor(obj)
	if err != nil {
		return nil, err
	}
	serviceName := meta.GetLabels()["rio.cattle.io/service"]
	if serviceName == "" {
		return nil, nil
	}
	return []relatedresource.Key{
		{
			Namespace: namespace,
			Name:      serviceName,
		},
	}, nil
}

type statusHandler struct {
	daemonSetCache   appsv1controller.DaemonSetCache
	statefulSetCache appsv1controller.StatefulSetCache
	deploymentCache  appsv1controller.DeploymentCache
	resolver         *endpoints.Resolver
}

func (s *statusHandler) handle(obj *riov1.Service, status riov1.ServiceStatus) (riov1.ServiceStatus, error) {
	owner, err := s.getOwner(obj)
	if err != nil {
		return status, err
	}

	endpoints, err := s.resolver.ServiceEndpoints(obj)
	if err != nil {
		return status, err
	}

	appEndpoints, err := s.resolver.AppEndpoints(obj)
	if err != nil {
		return status, err
	}

	status.Endpoints = endpoints
	status.AppEndpoints = appEndpoints
	status.ScaleStatus = toScaleStatus(owner)
	status.DeploymentReady = toReady(owner)
	return status, nil
}

func toScaleStatus(owner runtime.Object) *riov1.ScaleStatus {
	switch typed := owner.(type) {
	case *appsv1.Deployment:
		return &riov1.ScaleStatus{
			Available:   int(typed.Status.AvailableReplicas),
			Unavailable: int(typed.Status.UnavailableReplicas),
		}
	case *appsv1.DaemonSet:
		return &riov1.ScaleStatus{
			Available:   int(typed.Status.NumberAvailable),
			Unavailable: int(typed.Status.NumberUnavailable),
		}
	case *appsv1.StatefulSet:
		unavailable := typed.Status.Replicas - typed.Status.ReadyReplicas
		if unavailable < 0 {
			unavailable = 0
		}
		return &riov1.ScaleStatus{
			Available:   int(typed.Status.ReadyReplicas),
			Unavailable: int(unavailable),
		}
	default:
		return nil
	}
}

func toReady(owner runtime.Object) bool {
	switch typed := owner.(type) {
	case *appsv1.Deployment:
		return deployment.IsReady(&typed.Status)
	case *appsv1.DaemonSet:
		return typed.Status.NumberAvailable > 0
	case *appsv1.StatefulSet:
		return typed.Status.ReadyReplicas > 0
	}
	return false
}

func (s *statusHandler) getOwner(obj *riov1.Service) (runtime.Object, error) {
	deployment, err := s.deploymentCache.Get(obj.Namespace, obj.Name)
	if owned, err := isOwner(obj, deployment, err); err != nil {
		return nil, err
	} else if owned {
		return deployment, nil
	}

	daemonset, err := s.daemonSetCache.Get(obj.Namespace, obj.Name)
	if owned, err := isOwner(obj, daemonset, err); err != nil {
		return nil, err
	} else if owned {
		return daemonset, nil
	}

	statefulset, err := s.statefulSetCache.Get(obj.Namespace, obj.Name)
	if owned, err := isOwner(obj, statefulset, err); err != nil {
		return nil, err
	} else if owned {
		return statefulset, nil
	}

	return nil, nil
}

func isOwner(service *riov1.Service, object runtime.Object, err error) (bool, error) {
	if errors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	meta, err := meta.Accessor(object)
	if err != nil {
		return false, err
	}

	serviceName := meta.GetLabels()["rio.cattle.io/service"]
	return service.Name == serviceName, nil
}
