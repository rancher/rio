package ps

import (
	"strings"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/tables"
	"github.com/rancher/rio/cli/pkg/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ContainerData struct {
	ID        string
	Pod       *v1.Pod
	PodData   *tables.PodData
	Container *v1.Container
}

func ListFirstPod(ctx *clicontext.CLIContext, all bool, podOrServices ...string) (*tables.PodData, error) {
	cds, err := ListPods(ctx, all, podOrServices...)
	if len(cds) == 0 {
		return nil, err
	}
	return &cds[0], err
}

func ListPods(ctx *clicontext.CLIContext, all bool, podOrServices ...string) ([]tables.PodData, error) {
	var result []tables.PodData

	var pods []types.Resource
	var services []types.Resource

	for _, name := range podOrServices {
		r, err := lookup.Lookup(ctx, name, types.PodType, types.ServiceType)
		if err != nil {
			return nil, err
		}
		switch r.Type {
		case types.PodType:
			pods = append(pods, r)
		case types.ServiceType:
			services = append(services, r)
		}
	}

	for _, pod := range pods {
		containerName, _ := lookup.ParseContainer(ctx.DefaultStackName, pod.LookupName)
		pod := pod.Object.(*v1.Pod)
		podData, ok := toPodData(ctx, all, pod, containerName.ContainerName)
		if ok {
			result = append(result, podData)
		}
	}

	if len(pods) > 0 && len(services) == 0 {
		return result, nil
	}

	podList, err := ctx.Core.Pods("").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for i := range podList.Items {
		podData, ok := toPodData(ctx, all, &podList.Items[i], "")
		if !ok {
			continue
		}

		if len(services) == 0 {
			result = append(result, podData)
			continue
		}

		for _, service := range services {
			if service.Name == podData.Service.ResourceName && service.Namespace == podData.Service.StackName {
				result = append(result, podData)
				break
			}
		}
	}

	return result, nil
}

func toPodData(ctx *clicontext.CLIContext, all bool, pod *v1.Pod, containerName string) (tables.PodData, bool) {
	stackScoped := lookup.StackScopedFromLabels(ctx.DefaultStackName, pod.Labels)
	projectID := pod.Labels["rio.cattle.io/project"]

	podData := tables.PodData{
		Pod:     pod,
		Service: &stackScoped,
		Managed: projectID != "",
	}

	lookupName := stackScoped.ResourceName + "-"
	if strings.HasPrefix(pod.Name, lookupName) {
		podData.Name = strings.TrimPrefix(pod.Name, lookupName)
	} else {
		podData.Name = pod.Name
	}

	if !all && (podData.Name == "" || !podData.Managed || projectID != ctx.Namespace) {
		return podData, false
	}

	containers := append(pod.Spec.Containers, pod.Spec.InitContainers...)
	for _, container := range containers {
		if containerName == "" || container.Name == containerName {
			podData.Containers = append(podData.Containers, container)
		}
	}

	return podData, true
}

func (p *Ps) containers(ctx *clicontext.CLIContext) error {
	pds, err := ListPods(ctx, p.A_All, ctx.CLI.Args()...)
	if err != nil {
		return err
	}

	writer := tables.NewContainer(ctx)
	defer writer.TableWriter().Close()

	for _, pd := range pds {
		for _, container := range pd.Containers {
			writer.TableWriter().Write(ContainerData{
				ID:        pd.Name,
				PodData:   &pd,
				Container: &container,
			})
		}
	}

	return writer.TableWriter().Err()
}
