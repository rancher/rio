package riofile

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/rancher/rio/cli/pkg/table"

	"github.com/rancher/wrangler/pkg/crd"

	"github.com/rancher/mapper"
	"github.com/rancher/mapper/convert"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/template"
	"github.com/rancher/wrangler/pkg/yaml"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	schema = mapper.NewSchemas()
)

type Riofile struct {
	Services   map[string]riov1.Service
	Configs    map[string]v1.ConfigMap
	Kubernetes []runtime.Object
	CRD        []v1beta1.CustomResourceDefinition
}

type kubernetes struct {
	NamespacedCustomResourceDefintions []string `json:"namespacedCustomResourceDefinitions,omitempty"`
	CustomResourceDefintions           []string `json:"customResourceDefinitions,omitempty"`
	Manifest                           string   `json:"manifest,omitempty"`
}

type riofile struct {
	Services   map[string]riov1.Service `json:"services,omitempty"`
	Configs    map[string]v1.ConfigMap  `json:"configs,omitempty"`
	Kubernetes kubernetes               `json:"kubernetes,omitempty"`
}

func (r *Riofile) Objects() (result []runtime.Object) {
	for _, s := range r.CRD {
		copy := s
		result = append(result, &copy)
	}
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

	return
}

func ParseFrom(services map[string]riov1.Service, configs map[string]v1.ConfigMap) ([]byte, error) {
	rf := riofile{
		Services: services,
		Configs:  configs,
	}
	rawdata, err := json.Marshal(rf)
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{}
	if err := json.Unmarshal(rawdata, &data); err != nil {
		return nil, err
	}
	schema.Schema("Riofile").Mapper.FromInternal(data)
	result, err := table.FormatYAML(data)
	if err != nil {
		return nil, err
	}
	return []byte(result), nil
}

func Parse(reader io.Reader, answers template.AnswerCallback) (*Riofile, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	t := template.Template{
		Content: bytes,
	}

	content, err := t.Parse(answers)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{}
	if err := yaml.Unmarshal(content, &data); err != nil {
		return nil, err
	}

	if err := schema.Schema("Riofile").Mapper.ToInternal(data); err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	rf := &riofile{}
	if err := convert.ToObj(data, rf); err != nil {
		return nil, err
	}

	return toRiofile(rf)
}

func toRiofile(rf *riofile) (*Riofile, error) {
	riofile := &Riofile{
		Services: map[string]riov1.Service{},
		Configs:  map[string]v1.ConfigMap{},
	}

	for k, v := range rf.Services {
		v.Name = k
		riofile.Services[k] = v
	}

	for k, v := range rf.Configs {
		v.Name = k
		riofile.Configs[k] = v
	}

	if rf.Kubernetes.Manifest != "" {
		objs, err := yaml.ToObjects(bytes.NewBufferString(rf.Kubernetes.Manifest))
		if err != nil {
			return nil, err
		}
		riofile.Kubernetes = objs
	}

	for _, crdSpec := range rf.Kubernetes.CustomResourceDefintions {
		riofile.CRD = append(riofile.CRD, crd.NonNamespacedType(crdSpec).ToCustomResourceDefinition())
	}

	for _, crdSpec := range rf.Kubernetes.NamespacedCustomResourceDefintions {
		riofile.CRD = append(riofile.CRD, crd.NamespacedType(crdSpec).ToCustomResourceDefinition())
	}

	return riofile, nil
}
