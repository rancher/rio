package tables

import (
	"github.com/rancher/rio/cli/pkg/table"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

func NewStack(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{id .Obj}}"},
		{"CREATED", "{{.Obj.CreationTimestamp | ago}}"},
		{"DETAIL", "{{.Obj | stackDetail}}"},
	}, cfg)

	writer.AddFormatFunc("stackDetail", stackDetail)

	return &tableWriter{
		writer: writer,
	}
}

func stackDetail(data interface{}) string {
	stack, ok := data.(*riov1.Stack)
	if !ok {
		return ""
	}

	for _, con := range stack.Status.Conditions {
		if con.Type == "DeployedStackRiofile" && con.Status != "True" {
			return con.Message
		}
	}
	return ""
}
