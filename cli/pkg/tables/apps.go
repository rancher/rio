package tables

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rancher/rio/cli/pkg/table"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type AppData struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	App       *riov1.App
	Revisions map[string]struct {
		Revision *riov1.Service
		Pods     []corev1.Pod
	}
}

func (a *AppData) DeepCopyObject() runtime.Object {
	ad := &AppData{
		TypeMeta:   a.TypeMeta,
		ObjectMeta: *a.ObjectMeta.DeepCopy(),
		App:        a.App.DeepCopy(),
		Revisions: map[string]struct {
			Revision *riov1.Service
			Pods     []corev1.Pod
		}{},
	}

	for k, v := range a.Revisions {
		ad.Revisions[k] = v
	}

	return ad
}

func NewApp(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{stackScopedName .Obj.App.Namespace .Obj.App.Name ``}}"},
		{"CREATED", "{{.Obj.App.CreationTimestamp | ago}}"},
		{"ENDPOINT", "{{.Obj.App.Status.Endpoints | array}}"},
		{"REVISIONS", "{{revisions .Obj}}"},
		{"SCALE", "{{appScale .Obj}}"},
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

func revisions(app *riov1.App) (result []string) {
	for _, rev := range app.Spec.Revisions {
		revStatus := app.Status.RevisionWeight[rev.Version]
		if revStatus.Weight == 0 {
			continue
		}
		result = append(result, rev.Version)
	}
	sort.Strings(result)
	return
}

func revisionsByVersion(app *riov1.App) map[string]riov1.Revision {
	result := map[string]riov1.Revision{}
	for _, rev := range app.Spec.Revisions {
		result[rev.Version] = rev
	}
	return result
}

func formatRevisions(obj interface{}) (string, error) {
	revs := revisions(obj.(*AppData).App)
	return strings.Join(revs, ","), nil
}

func formatAppScale(obj interface{}) (string, error) {
	var (
		revMap = revisionsByVersion(obj.(*AppData).App)
		result []string
	)
	for _, version := range revisions(obj.(*AppData).App) {
		rev := revMap[version]
		var s *int
		if rev.Scale != 0 {
			s = &rev.Scale
		}
		scale, err := FormatScale(s, rev.ScaleStatus)
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

	for _, version := range revisions(app) {
		weight := app.Status.RevisionWeight[version].Weight
		ret := fmt.Sprintf("%v%%", weight)
		result = append(result, ret)
	}
	return strings.Join(result, ","), nil
}

func waitingOnBuild(svc *riov1.Service) bool {
	if svc.SystemSpec != nil {
		return false
	}

	if svc.Spec.Image == "" && len(svc.Spec.Sidecars) == 0 {
		return true
	}

	for _, container := range svc.Spec.Sidecars {
		if container.Image == "" {
			return true
		}
	}

	return false
}

func formatAppDetail(obj interface{}) (string, error) {
	appData := obj.(*AppData)
	buffer := strings.Builder{}

	versions := revisionsByVersion(appData.App)
	for _, name := range revisions(appData.App) {
		svc, ok := appData.Revisions[name]
		if !ok {
			continue
		}

		rev := versions[name]
		if !rev.DeploymentReady && !svc.Revision.Spec.Global {
			if buffer.Len() > 0 {
				buffer.WriteString("; ")
			}
			buffer.WriteString(name + ": ")
			pd, err := podsDetail(svc.Pods)
			if err != nil {
				return "", err
			}
			if pd == "" {
				pd = "not ready"
			}
			buffer.WriteString(pd)
		}

		if waitingOnBuild(svc.Revision) {
			if riov1.ServiceConditionImageReady.IsFalse(svc.Revision) {
				if buffer.Len() > 0 {
					buffer.WriteString("; ")
				}
				buffer.WriteString(name)
				buffer.WriteString(" build failed: ")
				buffer.WriteString(riov1.ServiceConditionImageReady.GetMessage(svc.Revision))
			} else if !riov1.ServiceConditionImageReady.IsTrue(svc.Revision) {
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
