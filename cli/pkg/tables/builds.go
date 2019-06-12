package tables

import (
	"fmt"

	"github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/rancher/rio/cli/pkg/table"
)

func NewBuild(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"Name", "{{stackScopedName .Obj.Namespace .Obj.Name ``}}"},
		{"SERVICE", "{{.Obj | findService}}"},
		{"REVISION", "{{.Obj | findRevision}}"},
		{"CREATED", "{{.Obj.CreationTimestamp | ago}}"},
		{"SUCCEED", "{{ .Obj | succeed }}"},
		{"REASON", "{{ .Obj | reason }}"},
	}, cfg)

	writer.AddFormatFunc("stackScopedName", table.FormatStackScopedName(cfg.GetSetNamespace()))
	writer.AddFormatFunc("findService", findService)
	writer.AddFormatFunc("findRevision", findRevision)
	writer.AddFormatFunc("succeed", findSucceed)
	writer.AddFormatFunc("reason", findReason)
	return &tableWriter{
		writer: writer,
	}
}

func findSucceed(data interface{}) (string, error) {
	b, ok := data.(*v1alpha1.Build)
	if !ok {
		return "", nil
	}
	cond := b.Status.GetCondition(v1alpha1.BuildSucceeded)
	if cond == nil {
		return "", nil
	}
	return string(cond.Status), nil
}

func findReason(data interface{}) (string, error) {
	b, ok := data.(*v1alpha1.Build)
	if !ok {
		return "", nil
	}
	cond := b.Status.GetCondition(v1alpha1.BuildSucceeded)
	if cond == nil {
		return "", nil
	}
	return string(cond.Reason), nil
}

func findService(data interface{}) (string, error) {
	m, ok := data.(*v1alpha1.Build)
	if !ok {
		return "", nil
	}
	name := m.Labels["service-name"]
	namespace := m.Labels["service-namespace"]
	return fmt.Sprintf("%s/%s", namespace, name), nil
}

func findRevision(data interface{}) (string, error) {
	m, ok := data.(*v1alpha1.Build)
	if !ok {
		return "", nil
	}
	return m.Spec.Source.Git.Revision, nil
}
