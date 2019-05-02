package tables

import (
	"fmt"

	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/table"
	v1 "k8s.io/api/core/v1"
)

type PodData struct {
	Name       string
	Managed    bool
	Service    *lookup.StackScoped
	Pod        *v1.Pod
	Containers []v1.Container
}

func NewContainer(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{containerName .PodData .Container}}"},
		{"IMAGE", "Container.Image"},
		{"CREATED", "{{.PodData.Pod.CreationTimestamp | ago}}"},
		{"NODE", "PodData.Pod.Spec.NodeName"},
		{"IP", "PodData.Pod.Status.PodIP"},
		{"STATE", "Container | toJson | state"},
		{"DETAIL", "Container | toJson | transitioning"},
	}, cfg)

	writer.AddFormatFunc("containerName", containerName)
	return &tableWriter{
		writer: writer,
	}
}

func containerName(obj, obj2 interface{}) (string, error) {
	podData, _ := obj.(*PodData)
	container, _ := obj2.(*v1.Container)

	if !podData.Managed {
		return fmt.Sprintf("%s/%s", podData.Name, container.Name), nil
	}

	pc := lookup.ParsedContainer{
		Service:       *podData.Service,
		ContainerName: container.Name,
		PodName:       podData.Name,
	}
	return pc.String(), nil
}
