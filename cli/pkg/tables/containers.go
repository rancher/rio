package tables

import (
	"fmt"
	"strings"

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

func NewPods(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{stackScopedName .PodData.Pod.Namespace .PodData.Pod.Name ``}}"},
		{"CREATED", "{{.PodData.Pod.CreationTimestamp | ago}}"},
		{"SERVICE", "{{.PodData.Service.ServiceName}}"},
		{"REVISION", "{{.PodData.Service.Version}}"},
		{"READY", "{{.PodData.Pod | podReady}}"},
		{"NODE", "PodData.Pod.Spec.NodeName"},
		{"IP", "PodData.Pod.Status.PodIP"},
		{"STATE", "PodData.Pod.Status.Phase"},
		{"DETAIL", "{{.PodData.Pod | podDetail}}"},
	}, cfg)

	writer.AddFormatFunc("stackScopedName", table.FormatStackScopedName(cfg.GetSetNamespace()))
	writer.AddFormatFunc("podReady", podReady)
	writer.AddFormatFunc("podDetail", podDetail)
	return &tableWriter{
		writer: writer,
	}
}

func NewContainer(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{containerName .PodData .Container}}"},
		{"IMAGE", "Container.Image | imageName"},
		{"CREATED", "{{.PodData.Pod.CreationTimestamp | ago}}"},
		{"NODE", "PodData.Pod.Spec.NodeName"},
		{"IP", "PodData.Pod.Status.PodIP"},
		{"STATE", "{{containerState .PodData.Pod .Container}}"},
		{"DETAIL", "{{containerDetail .PodData.Pod .Container}}"},
	}, cfg)

	writer.AddFormatFunc("containerName", containerName(cfg.GetSetNamespace()))
	writer.AddFormatFunc("imageName", imageName)
	writer.AddFormatFunc("containerState", containerState)
	writer.AddFormatFunc("containerDetail", containerDetail)
	return &tableWriter{
		writer: writer,
	}
}

func containerState(obj, obj2 interface{}) (string, error) {
	pod, _ := obj.(*v1.Pod)
	container, _ := obj2.(*v1.Container)

	for _, containerStatus := range append(pod.Status.ContainerStatuses, pod.Status.InitContainerStatuses...) {
		if containerStatus.Name == container.Name {
			if containerStatus.State.Running != nil {
				return "Running", nil
			}
			if containerStatus.State.Waiting != nil {
				return "Pending", nil
			}
			if containerStatus.State.Terminated != nil {
				if containerStatus.State.Terminated.ExitCode == 0 {
					return "Completed", nil
				}
				return "NotCompleted", nil
			}
		}
	}
	return "", nil
}

func containerDetail(obj, obj2 interface{}) (string, error) {
	pod, _ := obj.(*v1.Pod)
	container, _ := obj2.(*v1.Container)

	for _, containerStatus := range append(pod.Status.ContainerStatuses, pod.Status.InitContainerStatuses...) {
		if containerStatus.Name == container.Name {
			if containerStatus.State.Running != nil {
				return "", nil
			}
			if containerStatus.State.Waiting != nil {
				return fmt.Sprintf("%s: %s", containerStatus.State.Waiting.Reason, containerStatus.State.Waiting.Message), nil
			}

			if containerStatus.State.Terminated != nil && containerStatus.State.Terminated.ExitCode != 0 {
				return fmt.Sprintf("exit code: %v", containerStatus.State.Terminated.ExitCode), nil
			}
		}
	}
	return "", nil
}

func podDetail(obj interface{}) (string, error) {
	pod, _ := obj.(*v1.Pod)
	output := strings.Builder{}
	for _, con := range append(pod.Status.ContainerStatuses, pod.Status.InitContainerStatuses...) {
		if con.State.Waiting != nil && con.State.Waiting.Reason != "" {
			output.WriteString(";")
			output.WriteString(fmt.Sprintf("%s: %s", con.State.Waiting.Reason, con.State.Waiting.Message))
		}

		if con.State.Terminated != nil && con.State.Terminated.ExitCode != 0 {
			output.WriteString(";")
			if con.State.Terminated.Message == "" {
				con.State.Terminated.Message = "exit code not zero"
			}
			output.WriteString(fmt.Sprintf("%s: %s, exit code: %v", con.State.Terminated.Reason, con.State.Terminated.Message, con.State.Terminated.ExitCode))
		}
	}
	return strings.Trim(output.String(), ";"), nil
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

func imageName(obj interface{}) (string, error) {
	name, _ := obj.(string)
	return strings.TrimPrefix(name, "localhost:5442/"), nil
}

func containerName(defaultNamespace string) func(obj, obj2 interface{}) (string, error) {
	return func(obj, obj2 interface{}) (s string, e error) {
		podData, _ := obj.(*PodData)
		container, _ := obj2.(*v1.Container)

		if podData.Pod.Namespace == defaultNamespace {
			return fmt.Sprintf("%s/%s", podData.Pod.Name, container.Name), nil
		}

		return fmt.Sprintf("%s/%s/%s", podData.Pod.Namespace, podData.Pod.Name, container.Name), nil
	}
}
