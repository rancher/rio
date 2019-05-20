package tables

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rancher/rio/cli/pkg/table"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

func NewApp(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{stackScopedName .Obj.Namespace .Obj.Name ``}}"},
		{"ENDPOINT", "{{.Obj.Status.Endpoints | array}}"},
		{"REVISIONS", "{{revisions .Obj.Spec.Revisions}}"},
		{"SCALE", "{{appScale .Obj.Spec.Revisions}}"},
		{"WEIGHT", "{{weightVersioned .Obj}}"},
	}, cfg)

	writer.AddFormatFunc("revisions", formatRevisions)
	writer.AddFormatFunc("appScale", formatAppScale)
	writer.AddFormatFunc("weightVersioned", formatWeightGraph)
	writer.AddFormatFunc("stackScopedName", table.FormatStackScopedName(cfg.GetSetNamespace()))

	return &tableWriter{
		writer: writer,
	}
}

func revisions(obj interface{}) (result []string) {
	revs := obj.([]riov1.Revision)
	for _, rev := range revs {
		if rev.AdjustedWeight == 0 && rev.Scale == 0 {
			continue
		}
		result = append(result, rev.Version)
	}
	sort.Strings(result)
	return
}

func revisionsByVersion(obj interface{}) map[string]riov1.Revision {
	result := map[string]riov1.Revision{}
	for _, rev := range obj.([]riov1.Revision) {
		result[rev.Version] = rev
	}
	return result
}

func formatRevisions(obj interface{}) (string, error) {
	revs := revisions(obj)
	return strings.Join(revs, ","), nil
}

func formatAppScale(obj interface{}) (string, error) {
	var (
		revMap = revisionsByVersion(obj)
		result []string
	)
	for _, version := range revisions(obj) {
		rev := revMap[version]
		scale, err := FormatScale(rev.Scale, rev.ScaleStatus, nil)
		if err != nil {
			return "", nil
		}
		result = append(result, scale)
	}
	return strings.Join(result, ","), nil
}

func formatWeightGraph(obj interface{}) (string, error) {
	var (
		app    = obj.(*riov1.App)
		result []string
	)

	for _, version := range revisions(app.Spec.Revisions) {
		weight := app.Status.RevisionWeight[version].Weight
		ret := fmt.Sprintf("%v%%", weight)
		result = append(result, ret)
	}
	return strings.Join(result, ","), nil
}
