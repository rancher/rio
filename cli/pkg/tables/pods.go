package tables

import (
	"fmt"

	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/pkg/controllers/pkg"
	v1 "k8s.io/api/core/v1"
)

func NewPods(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{id .Obj}}"},
		{"CREATED", "{{.Obj.CreationTimestamp | ago}}"},
		{"READY", "{{.Obj | podReady}}"},
		{"NODE", "{{.Obj.Spec.NodeName}}"},
		{"IP", "{{.Obj.Status.PodIP}}"},
		{"STATE", "{{.Obj.Status.Phase}}"},
		{"DETAIL", "{{.Obj | podDetail}}"},
	}, cfg)

	writer.AddFormatFunc("podReady", podReady)
	writer.AddFormatFunc("podDetail", podDetail)
	return &tableWriter{
		writer: writer,
	}
}

func podDetail(obj interface{}) string {
	pod, _ := obj.(*v1.Pod)
	return pkg.PodDetail(pod)
}

func podReady(obj interface{}) (string, error) {
	podData, _ := obj.(*v1.Pod)
	ready := 0
	total := 0
	for _, con := range podData.Status.ContainerStatuses {
		if con.Ready {
			ready++
		}
		total++
	}
	return fmt.Sprintf("%v/%v", ready, total), nil
}
