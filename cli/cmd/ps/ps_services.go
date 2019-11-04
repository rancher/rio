package ps

import (
	"strings"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/tables"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ServiceData struct {
	Name       string
	Service    *riov1.Service
	Deployment *appsv1.Deployment
	DaemonSet  *appsv1.DaemonSet
	Namespace  string
	Pods       []*corev1.Pod
}

func (s ServiceData) Object() runtime.Object {
	if s.Deployment != nil {
		return s.Deployment
	} else if s.DaemonSet != nil {
		return s.DaemonSet
	}
	return s.Service
}

func findOwner(pod *corev1.Pod) string {
	mapKey := pod.Labels["rio.cattle.io/service"]
	if mapKey != "" {
		return mapKey
	}

	for _, ownerRef := range pod.OwnerReferences {
		if ownerRef.Kind == "DaemonSet" {
			return "ds/" + ownerRef.Name
		} else if ownerRef.Kind == "ReplicaSet" {
			idx := strings.LastIndex(ownerRef.Name, "-")
			if idx > 0 {
				return "deploy/" + ownerRef.Name[:idx]
			}
		}
	}

	return ""
}

func (p *Ps) services(ctx *clicontext.CLIContext) error {
	ds, err := ctx.List(clitypes.DaemonSetType)
	if err != nil {
		return err
	}

	deploys, err := ctx.List(clitypes.DeploymentType)
	if err != nil {
		return err
	}

	services, err := ctx.List(clitypes.ServiceType)
	if err != nil {
		return err
	}

	pods, err := ctx.List(clitypes.PodType)
	if err != nil {
		return err
	}

	podMap := map[string][]*corev1.Pod{}
	for _, obj := range pods {
		pod := obj.(*corev1.Pod)
		mapKey := findOwner(pod)
		if mapKey != "" {
			mapKey = pod.Namespace + "/" + mapKey
			podMap[mapKey] = append(podMap[mapKey], pod)
		}
	}

	var output []tables.Object

	for _, ds := range ds {
		ds := ds.(*appsv1.DaemonSet)
		if ds.Spec.Template.Labels["rio.cattle.io/service"] != "" {
			continue
		}
		output = append(output, ServiceData{
			DaemonSet: ds,
			Namespace: ds.Namespace,
			Pods:      podMap[ds.Namespace+"/ds/"+ds.Name],
		})
	}

	for _, deploy := range deploys {
		deploy := deploy.(*appsv1.Deployment)
		if deploy.Spec.Template.Labels["rio.cattle.io/service"] != "" {
			continue
		}
		output = append(output, ServiceData{
			Deployment: deploy,
			Namespace:  deploy.Namespace,
			Pods:       podMap[deploy.Namespace+"/deploy/"+deploy.Name],
		})
	}

	for _, service := range services {
		if service.(*riov1.Service).Spec.Template && !p.A_All {
			continue
		}
		output = append(output, ServiceData{
			Service:   service.(*riov1.Service),
			Namespace: service.(*riov1.Service).Namespace,
			Pods:      podMap[service.(*riov1.Service).Namespace+"/"+service.(*riov1.Service).Name],
		})
	}

	writer := tables.NewService(ctx)
	defer writer.TableWriter().Close()
	return writer.WriteObjects(output)
}
