package ps

import (
	"strings"

	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/kv"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/cli/server"
	spaceclient "github.com/rancher/rio/types/client/space/v1beta1"
	"github.com/urfave/cli"
)

var (
	ignoreNames = map[string]bool{
		"istio-proxy":      true,
		"istio-init":       true,
		"enable-core-dump": true,
	}
)

type ContainerData struct {
	ID        string
	Pod       *spaceclient.Pod
	Container *spaceclient.Container
}

func ListFirstPod(c *spaceclient.Client, all bool, specificContainerName string, criteria ...string) (*ContainerData, error) {
	cds, err := ListPods(c, all, specificContainerName, criteria...)
	if len(cds) == 0 {
		return nil, err
	}
	return cds[0], err
}

func ListPods(c *spaceclient.Client, all bool, specificContainerName string, criteria ...string) ([]*ContainerData, error) {
	var result []*ContainerData

	pods, err := c.Pod.List(util.DefaultListOpts())
	if err != nil {
		return nil, err
	}

	filters := filter(criteria)

	for i, pod := range pods.Data {
		containers := append(pod.Containers, pod.InitContainers...)
		for j, container := range containers {
			serviceName := pod.Labels["rio.cattle.io/service"]
			if !all && (serviceName == "" || ignoreNames[container.Name]) {
				continue
			}

			cd := &ContainerData{
				Pod:       &pods.Data[i],
				Container: &containers[j],
			}

			cd.ID, _ = containerName(cd.Pod, cd.Container)

			if cd.Pod.Transitioning == "error" && cd.Container.TransitioningMessage == "" {
				cd.Container.State = cd.Pod.State
				cd.Container.TransitioningMessage = cd.Pod.TransitioningMessage
			}

			if !shouldSkip(filters, cd) && (specificContainerName == "" || specificContainerName == cd.Container.Name) {
				result = append(result, cd)
			}
		}
	}

	return result, nil
}

func (p *Ps) containers(app *cli.Context, ctx *server.Context) error {
	c, err := ctx.SpaceClient()
	if err != nil {
		return err
	}

	cds, err := ListPods(c, p.A_All, "", app.Args()...)
	if err != nil {
		return err
	}

	writer := table.NewWriter([][]string{
		{"NAME", "{{containerName .Pod .Container}}"},
		{"IMAGE", "Container.Image"},
		{"CREATED", "{{.Pod.Created | ago}}"},
		{"NODE", "Pod.NodeName"},
		{"IP", "Pod.PodIP"},
		{"STATE", "Container.State"},
		{"DETAIL", "Container.TransitioningMessage"},
	}, app)
	defer writer.Close()

	writer.AddFormatFunc("containerName", containerName)

	for _, cd := range cds {
		writer.Write(cd)
	}

	return writer.Err()
}

func shouldSkip(filters []func(*ContainerData) bool, cd *ContainerData) bool {
	if len(filters) == 0 {
		return false
	}

	for _, f := range filters {
		if f(cd) {
			return false
		}
	}

	return true
}

func filter(args []string) []func(*ContainerData) bool {
	var result []func(*ContainerData) bool

	for _, arg := range args {
		con, ok := lookup.ParseContainerName(arg)
		if ok {
			result = append(result, func(cd *ContainerData) bool {
				return cd.Pod.Name == con.PodName && cd.Container.Name == con.ContainerName
			})
			continue
		}

		svc := lookup.ParseServiceName(arg)
		prefix := svc.String() + "/"
		result = append(result, func(cd *ContainerData) bool {
			name, err := containerName(cd.Pod, cd.Container)
			if err != nil {
				return false
			}
			return strings.HasPrefix(name, prefix)
		})

		ns, service := kv.Split(arg, ":")
		result = append(result, func(cd *ContainerData) bool {
			return cd.Pod.Labels["rio.cattle.io/namespace"] == ns && cd.Pod.Labels["rio.cattle.io/service"] == service
		})
	}

	return result
}

func containerName(obj, obj2 interface{}) (string, error) {
	pod, _ := obj.(*spaceclient.Pod)
	container, _ := obj2.(*spaceclient.Container)

	return lookup.ParsedContainer{
		Service:       lookup.ParseServiceNameFromLabels(pod.Labels),
		ContainerName: container.Name,
		PodName:       pod.Name,
	}.String(), nil
}
