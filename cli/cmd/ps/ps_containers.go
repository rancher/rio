package ps

import (
	"fmt"
	"strings"

	"github.com/rancher/norman/types"
	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/clientcfg"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/pkg/settings"
	projectclient "github.com/rancher/rio/types/client/project/v1"
	client "github.com/rancher/rio/types/client/rio/v1"
)

type PodData struct {
	ID         string
	Name       string
	Managed    bool
	Service    *lookup.StackScoped
	Pod        *projectclient.Pod
	Containers []projectclient.Container
}

type ContainerData struct {
	ID        string
	Pod       *projectclient.Pod
	PodData   *PodData
	Container *projectclient.Container
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

	c, err := ctx.ClusterClient()
	if err != nil {
		return nil, err
	}

	if w.ID != settings.RioSystemNamespace {
		all = false
	}

	var pods []*types.NamedResource
	var services []*types.NamedResource

	for _, name := range podOrServices {
		r, err := lookup.Lookup(ctx, name, projectclient.PodType, client.ServiceType)
		if err != nil {
			return nil, err
		}
		switch r.Type {
		case projectclient.PodType:
			pods = append(pods, r)
		case client.ServiceType:
			services = append(services, r)
		}
	}

	for _, pod := range pods {
		pod, err := c.Pod.ByID(pod.ID)
		if err != nil {
			return nil, err
		}

		podData, ok := toPodData(w, all, pod)
		if ok {
			result = append(result, podData)
		}
	}

	if len(pods) > 0 && len(services) == 0 {
		return result, nil
	}

	podList, err := c.Pod.List(util.DefaultListOpts())
	if err != nil {
		return nil, err
	}

	for i := range podList.Data {
		podData, ok := toPodData(w, all, &podList.Data[i])
		if !ok {
			continue
		}

		if len(services) == 0 {
			result = append(result, podData)
			continue
		}

		for _, service := range services {
			if service.ID == podData.Service.ResourceID {
				result = append(result, podData)
				break
			}
		}
	}

	return result, nil
}

func toPodData(w *clientcfg.Project, all bool, pod *projectclient.Pod) (PodData, bool) {
	stackScoped := lookup.StackScopedFromLabels(w, pod.Labels)
	projectID := pod.Labels["rio.cattle.io/project"]

	podData := PodData{
		ID:      pod.ID,
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

	if !all && (podData.Name == "" || !podData.Managed || projectID != w.ID) {
		return podData, false
	}

	containers := append(pod.Containers, pod.InitContainers...)
	for _, container := range containers {
		if podData.Pod.Transitioning == "error" && container.TransitioningMessage == "" {
			container.State = podData.Pod.State
			container.TransitioningMessage = podData.Pod.TransitioningMessage
		}

		podData.Containers = append(podData.Containers, container)
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
		{"CREATED", "{{.PodData.Pod.Created | ago}}"},
		{"NODE", "PodData.Pod.NodeName"},
		{"IP", "PodData.Pod.PodIP"},
		{"STATE", "Container.State"},
		{"DETAIL", "Container.TransitioningMessage"},
	}, ctx)
	defer writer.Close()

	writer.AddFormatFunc("containerName", containerName)

	for _, pd := range pds {
		for _, container := range pd.Containers {
			writer.Write(ContainerData{
				ID:        pd.ID,
				PodData:   &pd,
				Container: &container,
			})
		}
	}

	return writer.Err()
}

func containerName(obj, obj2 interface{}) (string, error) {
	podData, _ := obj.(*PodData)
	container, _ := obj2.(*projectclient.Container)

	if !podData.Managed {
		return fmt.Sprintf("%s/%s", strings.Split(podData.ID, ":")[1], container.Name), nil
	}

	pc := lookup.ParsedContainer{
		Service:       *podData.Service,
		ContainerName: container.Name,
		PodName:       podData.Name,
	}
	return pc.String(), nil
}
