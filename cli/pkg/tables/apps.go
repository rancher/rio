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
		{"SCALE", "{{appScale .Obj.Spec.Revisions}}"},
		{"WEIGHT", "{{weightVersioned .Obj}}"},
	}, cfg)

	writer.AddFormatFunc("appScale", formatAppScale)
	writer.AddFormatFunc("weightVersioned", formatWeightGraph)
	writer.AddFormatFunc("stackScopedName", table.FormatStackScopedName(cfg.GetSetNamespace()))

	return &tableWriter{
		writer: writer,
	}
}

func formatAppScale(obj interface{}) (string, error) {
	revisions := obj.([]riov1.Revision)
	builder := &strings.Builder{}
	for _, rev := range revisions {
		scale, err := FormatScale(rev.Scale, rev.ScaleStatus, nil)
		if err != nil {
			return "", nil
		}
		builder.WriteString(fmt.Sprintf("%v/%v; ", rev.Version, scale))
	}
	return strings.Trim(builder.String(), "; "), nil
}

func formatWeightGraph(obj interface{}) (string, error) {
	app := obj.(*riov1.App)
	builder := &strings.Builder{}
	var versions []string
	versionMap := make(map[string]int)
	for _, rev := range app.Spec.Revisions {
		versions = append(versions, rev.Version)
		versionMap[rev.Version] = rev.AdjustedWeight
	}
	sort.Strings(versions)
	for _, version := range versions {
		weight := app.Status.RevisionWeight[version].Weight
		desiredWeight := versionMap[version]
		ret := fmt.Sprintf("%v/%v%s", version, weight, "%")
		if weight != desiredWeight {
			ret += fmt.Sprintf("(%v%s)", desiredWeight, "%")
		}
		builder.WriteString(ret)
		builder.WriteString("; ")
	}
	return strings.Trim(builder.String(), "; "), nil
}
