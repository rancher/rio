package tables

import "github.com/rancher/rio/cli/pkg/table"

func NewImage(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"REPO", "{{.Repo}}"},
		{"TAG", "{{.Tag}}"},
		{"IMAGE", "{{.Image}}"},
	}, cfg)

	return &tableWriter{
		writer: writer,
	}
}
