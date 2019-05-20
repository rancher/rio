package tables

import (
	"fmt"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/rancher/rio/cli/pkg/table"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AppData struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	App       *riov1.App
	Revisions map[string]*riov1.Service
}

func (a *AppData) DeepCopyObject() runtime.Object {
	ad := &AppData{
		TypeMeta:   a.TypeMeta,
		ObjectMeta: *a.ObjectMeta.DeepCopy(),
		App:        a.App.DeepCopy(),
		Revisions:  map[string]*riov1.Service{},
	}

	for k, v := range a.Revisions {
		ad.Revisions[k] = v.DeepCopy()
	}

	return ad
}

func NewApp(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{stackScopedName .Obj.App.Namespace .Obj.App.Name ``}}"},
		{"CREATED", "{{.Obj.App.CreationTimestamp | ago}}"},
		{"ENDPOINT", "{{.Obj.App.Status.Endpoints | array}}"},
		{"REVISIONS", "{{revisions .Obj.App.Spec.Revisions}}"},
		{"SCALE", "{{appScale .Obj.App.Spec.Revisions}}"},
		{"WEIGHT", "{{weightVersioned .Obj.App}}"},
		{"DETAIL", "{{detail .Obj}}"},
	}, cfg)

	writer.AddFormatFunc("revisions", formatRevisions)
	writer.AddFormatFunc("appScale", formatAppScale)
	writer.AddFormatFunc("weightVersioned", formatWeightGraph)
	writer.AddFormatFunc("detail", formatAppDetail)
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

func formatAppDetail(obj interface{}) (string, error) {
	appData := obj.(*AppData)
	buffer := strings.Builder{}

	versions := revisionsByVersion(appData.App.Spec.Revisions)
	for _, name := range revisions(appData.App.Spec.Revisions) {
		rev := versions[name]
		if !rev.DeploymentReady {
			if buffer.Len() > 0 {
				buffer.WriteString("; ")
			}
			buffer.WriteString(name)
			buffer.WriteString(" NotReady")
		}

		svc, ok := appData.Revisions[name]
		if ok && svc.Spec.Image == "" {
			if riov1.ServiceConditionImageReady.IsFalse(svc) {
				if buffer.Len() > 0 {
					buffer.WriteString("; ")
				}
				buffer.WriteString(name)
				buffer.WriteString(" build failed: ")
				buffer.WriteString(riov1.ServiceConditionImageReady.GetMessage(svc))
			} else if !riov1.ServiceConditionImageReady.IsTrue(svc) {
				if buffer.Len() > 0 {
					buffer.WriteString("; ")
				}
				buffer.WriteString(name)
				buffer.WriteString(" waiting on build")
			}
		}
	}
	return buffer.String(), nil
}
