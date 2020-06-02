package tables

import (
	"fmt"

	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/cli/pkg/types"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/wrangler/pkg/condition"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
)

func NewBuild(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{id .Obj}}"},
		{"SERVICE/STACK", "{{.Obj | findService}}"},
		{"REVISION", "{{.Obj | findRevision}}"},
		{"CREATED", "{{.Obj.CreationTimestamp | ago}}"},
		{"SUCCEED", "{{ .Obj | succeed }}"},
		{"REASON", "{{ .Obj | reason }}"},
	}, cfg)

	writer.AddFormatFunc("findService", findService)
	writer.AddFormatFunc("findRevision", findRevision)
	writer.AddFormatFunc("succeed", findSucceed)
	writer.AddFormatFunc("reason", findReason)
	return &tableWriter{
		writer: writer,
	}
}

func findSucceed(data interface{}) (string, error) {
	b, ok := data.(*tektonv1alpha1.TaskRun)
	if !ok {
		return "", nil
	}
	return condition.Cond("Succeeded").GetStatus(b), nil
}

func findReason(data interface{}) (string, error) {
	b, ok := data.(*tektonv1alpha1.TaskRun)
	if !ok {
		return "", nil
	}
	return condition.Cond("Succeeded").GetReason(b), nil
}

func findService(data interface{}) (string, error) {
	m, ok := data.(*tektonv1alpha1.TaskRun)
	if !ok {
		return "", nil
	}
	if m.Labels[constants.ServiceLabel] != "" {
		name := m.Labels[constants.ServiceLabel]
		return fmt.Sprintf("%s/%s", types.ServiceType, name), nil
	} else if m.Labels[constants.StackLabel] != "" {
		name := m.Labels[constants.StackLabel]
		return fmt.Sprintf("%s/%s", types.StackType, name), nil
	}
	return "", nil
}

func findRevision(data interface{}) (string, error) {
	m, ok := data.(*tektonv1alpha1.TaskRun)
	if !ok {
		return "", nil
	}
	for _, param := range m.Spec.Resources.Inputs[0].PipelineResourceBinding.ResourceSpec.Params {
		if param.Name == "revision" {
			return param.Value, nil
		}
	}
	return "", nil
}
