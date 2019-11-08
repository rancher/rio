package ps

import (
	"math"
	"strings"

	webhookv1 "github.com/rancher/gitwatcher/pkg/apis/gitwatcher.cattle.io/v1"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/tables"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ServiceData struct {
	Name       string
	Service    *riov1.Service
	Weight     int
	Deployment *appsv1.Deployment
	DaemonSet  *appsv1.DaemonSet
	Namespace  string
	Pods       []*corev1.Pod
	GitWatcher *webhookv1.GitWatcher
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

	if p.W_Workloads {
		ds, err := ctx.List(clitypes.DaemonSetType)
		if err != nil {
			return err
		}

		deploys, err := ctx.List(clitypes.DeploymentType)
		if err != nil {
			return err
		}

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
	}

	for _, service := range services {
		if service.(*riov1.Service).Spec.Template && !p.A_All {
			continue
		}

		weight := 0
		if service.(*riov1.Service).Status.ComputedWeight != nil && *service.(*riov1.Service).Status.ComputedWeight > 0 {
			totalWeight := 0
			for _, subService := range services {
				if service.(*riov1.Service).Spec.App == subService.(*riov1.Service).Spec.App && subService.(*riov1.Service).Status.ComputedWeight != nil {
					totalWeight += *subService.(*riov1.Service).Status.ComputedWeight
				}
			}
			weight = int(math.Round((float64(*service.(*riov1.Service).Status.ComputedWeight) / float64(totalWeight)) / 0.01)) // round to nearest percent
		}

		var gitwatcher *webhookv1.GitWatcher
		gitwatchers, err := ctx.Gitwatcher.GitWatchers(service.(*riov1.Service).Namespace).List(metav1.ListOptions{})
		if err == nil {
			for _, gw := range gitwatchers.Items {
				if len(gw.OwnerReferences) > 0 && gw.OwnerReferences[0].UID == service.(*riov1.Service).UID {
					gitwatcher = &gw
					break
				}
			}
		}
		output = append(output, ServiceData{
			Service:    service.(*riov1.Service),
			Weight:     weight,
			Namespace:  service.(*riov1.Service).Namespace,
			Pods:       podMap[service.(*riov1.Service).Namespace+"/"+service.(*riov1.Service).Name],
			GitWatcher: gitwatcher,
		})
	}

	writer := tables.NewService(ctx)
	defer writer.TableWriter().Close()
	return writer.WriteObjects(output)
}
