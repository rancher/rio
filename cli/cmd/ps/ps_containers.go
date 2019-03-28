package ps

import (
	"fmt"
	"os"
	"strings"

	"github.com/rancher/rio/cli/pkg/mapper"

	"github.com/rancher/norman/pkg/kv"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/clientcfg"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/cli/pkg/types"
	"github.com/rancher/rio/pkg/settings"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodData struct {
	Name       string
	Managed    bool
	Service    *lookup.StackScoped
	Pod        *v1.Pod
	Containers []v1.Container
}

type ContainerData struct {
	ID        string
	Pod       *v1.Pod
	PodData   *PodData
	Container *v1.Container
}

func ListFirstPod(ctx *clicontext.CLIContext, all bool, podOrServices ...string) (*PodData, error) {
	cds, err := ListPods(ctx, all, podOrServices...)
	if len(cds) == 0 {
		return nil, err
	}
	return &cds[0], err
}

func ListPods(ctx *clicontext.CLIContext, all bool, podOrServices ...string) ([]PodData, error) {
	var result []PodData

	w, err := ctx.Project()
	if err != nil {
		return nil, err
	}

	cluster, err := ctx.Cluster()
	if err != nil {
		return nil, err
	}

	client, err := cluster.KubeClient()
	if err != nil {
		return nil, err
	}

	if w.Project.Name != settings.RioSystemNamespace {
		all = false
	}

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
		if len(strings.Split(pod.Name, "/")) != 2 {
			continue
		}
		podname, containername := kv.Split(pod.Name, "/")
		pod, err := client.Core.Pods(pod.Namespace).Get(podname, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		podData, ok := toPodData(w, all, pod, containername)
		if ok {
			result = append(result, podData)
		}
	}

	if len(pods) > 0 && len(services) == 0 {
		return result, nil
	}

	podList, err := client.Core.Pods("").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for i := range podList.Items {
		podData, ok := toPodData(w, all, &podList.Items[i], "")
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

func toPodData(w *clientcfg.Project, all bool, pod *v1.Pod, containerName string) (PodData, bool) {
	stackScoped := lookup.StackScopedFromLabels(w, pod.Labels)
	projectID := pod.Labels["rio.cattle.io/project"]

	podData := PodData{
		Pod:     pod,
		Service: &stackScoped,
		Managed: projectID != "",
	}

	lookupName := stackScoped.LookupName() + "-"
	if strings.HasPrefix(pod.Name, lookupName) {
		podData.Name = strings.TrimPrefix(pod.Name, lookupName)
	} else {
		podData.Name = pod.Name
	}

	if !all && (podData.Name == "" || !podData.Managed || projectID != w.Project.Name) {
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

	writer := table.NewWriter([][]string{
		{"NAME", "{{containerName .PodData .Container}}"},
		{"IMAGE", "Container.Image"},
		{"CREATED", "{{.PodData.Pod.CreationTimestamp | ago}}"},
		{"NODE", "PodData.Pod.Spec.NodeName"},
		{"IP", "PodData.Pod.Status.PodIP"},
		{"STATE", "Container | toJson | state"},
		{"DETAIL", "Container | toJson | transistioning"},
	}, ctx, os.Stdout)
	defer writer.Close()

	m := mapper.GenericStatusMapper
	writer.AddFormatFunc("containerName", containerName)
	writer.AddFormatFunc("state", m.FormatState)
	writer.AddFormatFunc("transistioning", m.FormatTransitionMessage)

	for _, pd := range pds {
		for _, container := range pd.Containers {
			writer.Write(ContainerData{
				ID:        pd.Name,
				PodData:   &pd,
				Container: &container,
			})
		}
	}

	return writer.Err()
}

func containerName(obj, obj2 interface{}) (string, error) {
	podData, _ := obj.(*PodData)
	container, _ := obj2.(*v1.Container)

	if !podData.Managed {
		return fmt.Sprintf("%s/%s", strings.Split(podData.Name, ":")[1], container.Name), nil
	}

	pc := lookup.ParsedContainer{
		Service:       *podData.Service,
		ContainerName: container.Name,
		PodName:       podData.Name,
	}
	return pc.String(), nil
}
