package riofile

import (
	"bytes"
	"strings"

	"github.com/rancher/norman/pkg/types/convert"
	"github.com/rancher/rio/cli/pkg/table"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/riofile/schema"
	"github.com/rancher/rio/pkg/template"
	"github.com/rancher/wrangler/pkg/gvk"
	"github.com/rancher/wrangler/pkg/yaml"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type Riofile struct {
	Services         map[string]riov1.Service
	Configs          map[string]v1.ConfigMap
	Routers          map[string]riov1.Router
	ExternalServices map[string]riov1.ExternalService
	Kubernetes       []runtime.Object
}

func (r *Riofile) Objects() (result []runtime.Object) {
	for _, s := range r.Kubernetes {
		result = append(result, s)
	}
	for _, s := range r.Configs {
		copy := s
		result = append(result, &copy)
	}
	for _, s := range r.Services {
		copy := s
		result = append(result, &copy)
	}
	for _, s := range r.ExternalServices {
		copy := s
		result = append(result, &copy)
	}
	for _, s := range r.Routers {
		copy := s
		result = append(result, &copy)
	}
	return
}

func RenderObject(object runtime.Object) ([]byte, error) {
	data, err := convert.EncodeToMap(object)
	if err != nil {
		return nil, err
	}

	kind := object.GetObjectKind().GroupVersionKind().Kind
	schema.Schema.Schema(kind).Mapper.FromInternal(data)

	result, err := table.FormatYAML(data)
	if err != nil {
		return nil, err
	}
	return []byte(result), nil
}

func Render(objects []runtime.Object) ([]byte, error) {
	rf := schema.ExternalRiofile{
		Services:         map[string]riov1.Service{},
		Configs:          map[string]v1.ConfigMap{},
		Routers:          map[string]riov1.Router{},
		ExternalServices: map[string]riov1.ExternalService{},
	}

	var other []runtime.Object

	for _, obj := range objects {
		switch obj.(type) {
		case *riov1.Service:
			svc := obj.(*riov1.Service)
			rf.Services[svc.Name] = *svc
		case *v1.ConfigMap:
			cm := obj.(*v1.ConfigMap)
			rf.Configs[cm.Name] = *cm
		case *riov1.Router:
			router := obj.(*riov1.Router)
			rf.Routers[router.Name] = *router
		case *riov1.ExternalService:
			es := obj.(*riov1.ExternalService)
			rf.ExternalServices[es.Name] = *es
		case runtime.Object:
			other = append(other, obj)
		}
	}

	if len(other) > 0 {
		bytes, err := yaml.ToBytes(other)
		if err != nil {
			return nil, err
		}
		rf.Kubernetes = &schema.Kubernetes{
			Manifest: string(bytes),
		}
	}

	data, err := convert.EncodeToMap(rf)
	if err != nil {
		return nil, err
	}

	schema.Schema.Schema("Riofile").Mapper.FromInternal(data)
	result, err := table.FormatYAML(data)
	if err != nil {
		return nil, err
	}
	return []byte(result), nil
}

func isK8SYaml(contents []byte) (bool, []runtime.Object, error) {
	objs, err := yaml.ToObjects(bytes.NewBuffer(contents))
	if err != nil {
		return false, nil, nil
	}
	if len(objs) > 0 &&
		objs[0].GetObjectKind().GroupVersionKind().Kind != "" {
		return true, objs, nil
	}
	return false, nil, nil
}

func Parse(contents []byte, answers template.AnswerCallback) (*Riofile, error) {
	k8s, objs, err := isK8SYaml(contents)
	if err != nil {
		return nil, err
	}

	if k8s {
		return &Riofile{
			Kubernetes: objs,
		}, nil
	}

	data, err := parseData(contents, answers)
	if err != nil {
		return nil, err
	}

	if err := schema.Schema.Schema("Riofile").Mapper.ToInternal(data); err != nil {
		return nil, err
	}

	rf := &schema.ExternalRiofile{}
	if err := convert.ToObj(data, rf); err != nil {
		return nil, err
	}

	return toRiofile(rf)
}

func parseData(contents []byte, answers template.AnswerCallback) (map[string]interface{}, error) {
	t := template.Template{
		Content: contents,
	}

	cont, err := t.Parse(answers)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{}
	if err := yaml.Unmarshal(cont, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func Update(originalObj runtime.Object, bytes []byte) (runtime.Object, error) {
	data, err := parseData(bytes, nil)
	if err != nil {
		return nil, err
	}

	kind, err := gvk.Get(originalObj)
	if err != nil {
		return nil, err
	}
	if err := schema.Schema.Schema(kind.Kind).Mapper.ToInternal(data); err != nil {
		return nil, err
	}

	originalData, err := convert.EncodeToMap(originalObj)
	if err != nil {
		return nil, err
	}

	annotations := make(map[string]interface{})
	labels := make(map[string]interface{})

	originalMeta, _ := originalData["metadata"].(map[string]interface{})
	originalAnno, _ := originalMeta["annotations"].(map[string]interface{})
	for k, v := range originalAnno {
		if strings.Contains(k, "rio.cattle.io") {
			annotations[k] = v
		}
	}
	originalLabel, _ := originalMeta["labels"].(map[string]interface{})
	for k, v := range originalLabel {
		if strings.Contains(k, "rio.cattle.io") {
			labels[k] = v
		}
	}

	meta, _ := data["metadata"].(map[string]interface{})
	modifiedAnno, _ := meta["annotations"].(map[string]interface{})
	modifiedLabels, _ := meta["labels"].(map[string]interface{})

	originalMeta["annotations"] = merge(annotations, modifiedAnno)
	originalMeta["labels"] = merge(labels, modifiedLabels)

	data["metadata"] = originalMeta
	data["status"] = originalData["status"]

	return &unstructured.Unstructured{
		Object: data,
	}, nil
}

func merge(labels1, labels2 map[string]interface{}) map[string]interface{} {
	mergedMap := map[string]interface{}{}

	for k, v := range labels1 {
		mergedMap[k] = v
	}
	for k, v := range labels2 {
		mergedMap[k] = v
	}
	return mergedMap
}

func toRiofile(rf *schema.ExternalRiofile) (*Riofile, error) {
	riofile := &Riofile{
		Services:         map[string]riov1.Service{},
		Configs:          map[string]v1.ConfigMap{},
		Routers:          map[string]riov1.Router{},
		ExternalServices: map[string]riov1.ExternalService{},
	}

	for k, v := range rf.Services {
		v.Name = k
		riofile.Services[k] = v
	}

	for k, v := range rf.Configs {
		v.Name = k
		riofile.Configs[k] = v
	}

	for k, v := range rf.Routers {
		v.Name = k
		riofile.Routers[k] = v
	}

	for k, v := range rf.ExternalServices {
		v.Name = k
		riofile.ExternalServices[k] = v
	}

	if rf.Kubernetes != nil {
		if rf.Kubernetes.Manifest != "" {
			objs, err := yaml.ToObjects(bytes.NewBufferString(rf.Kubernetes.Manifest))
			if err != nil {
				return nil, err
			}
			riofile.Kubernetes = objs
		}
	}

	return riofile, nil
}
