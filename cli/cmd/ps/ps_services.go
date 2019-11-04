package ps

import (
	"sort"

	webhookv1 "github.com/rancher/gitwatcher/pkg/apis/gitwatcher.cattle.io/v1"
	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/tables"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceData struct {
	Name       string
	Service    *riov1.Service
	Namespace  string
	Pods       []*corev1.Pod
	GitWatcher *webhookv1.GitWatcher
}

func (p *Ps) services(ctx *clicontext.CLIContext) error {
	namespace := ctx.GetSetNamespace()
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
		if pod.Labels["rio.cattle.io/service"] != "" {
			podMap[pod.Labels["rio.cattle.io/service"]] = append(podMap[pod.Labels["rio.cattle.io/service"]], pod)
		}
	}

	var output []ServiceData

	for _, service := range services {
		if service.(*riov1.Service).Spec.Template && !p.A_All {
			continue
		}
		allNamespace := namespace == ""
		id, err := util.GetID(service, allNamespace)
		if err != nil {
			return err
		}
		var gitwatcher *webhookv1.GitWatcher
		gitwatchers, err := ctx.Gitwatcher.GitWatchers(namespace).List(metav1.ListOptions{})
		if err == nil {
			for _, gw := range gitwatchers.Items {
				if len(gw.OwnerReferences) > 0 && gw.OwnerReferences[0].UID == service.(*riov1.Service).UID {
					gitwatcher = &gw
					break
				}
			}
		}
		output = append(output, ServiceData{
			Name:       id,
			Service:    service.(*riov1.Service),
			Namespace:  service.(*riov1.Service).Namespace,
			Pods:       podMap[service.(*riov1.Service).Name],
			GitWatcher: gitwatcher,
		})
	}

	sort.Slice(output, func(i, j int) bool {
		leftMeta, _ := meta.Accessor(output[i].Service)
		rightMeta, _ := meta.Accessor(output[j].Service)
		if leftMeta.GetNamespace() != rightMeta.GetNamespace() {
			return leftMeta.GetNamespace() < rightMeta.GetNamespace()
		}
		leftCreated := leftMeta.GetCreationTimestamp()
		return leftCreated.After(rightMeta.GetCreationTimestamp().Time)
	})

	writer := tables.NewService(ctx)
	defer writer.TableWriter().Close()
	for _, obj := range output {
		writer.TableWriter().Write(obj)
	}
	return writer.TableWriter().Err()
}
