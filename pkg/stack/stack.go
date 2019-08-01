package stack

import (
	"bytes"
	"encoding/json"
	"strings"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/riofile"
	"github.com/rancher/rio/pkg/template"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

type Stack struct {
	contents []byte
	answers  map[string]string
	objects  []runtime.Object
}

func NewStack(contents []byte, answers map[string]string) *Stack {
	return &Stack{
		contents: contents,
		answers:  answers,
	}
}

func (s *Stack) Questions() ([]v1.Question, error) {
	t := template.Template{
		Content: s.contents,
	}

	if err := t.Validate(); err != nil {
		return nil, err
	}

	return t.Questions()
}

func (s *Stack) Yaml(answers map[string]string) (string, error) {
	rf, err := riofile.Parse(bytes.NewBuffer(s.contents), template.AnswersFromMap(answers))
	if err != nil {
		return "", err
	}

	output := strings.Builder{}
	objs := rf.Objects()
	for _, obj := range objs {
		data, err := json.Marshal(obj)
		if err != nil {
			return "", err
		}
		r, err := yaml.JSONToYAML(data)
		if err != nil {

			return "", err
		}
		output.Write(r)
		output.WriteString("---\n")
	}
	return output.String(), nil
}

func (s *Stack) GetObjects() ([]runtime.Object, error) {
	if s.objects != nil {
		return s.objects, nil
	}

	rf, err := riofile.Parse(bytes.NewBuffer(s.contents), template.AnswersFromMap(s.answers))
	if err != nil {
		return nil, err
	}
	return rf.Objects(), nil
}

func (s *Stack) GetImageBuilds() (map[string]riov1.ImageBuild, error) {
	objs, err := s.GetObjects()
	if err != nil {
		return nil, err
	}

	buildConfig := make(map[string]riov1.ImageBuild)
	for _, obj := range objs {
		if svc, ok := obj.(*riov1.Service); ok {
			if svc.Spec.Image == "" {
				buildConfig[svc.Name] = *svc.Spec.Build
			}
		}
	}
	return buildConfig, nil
}

func (s *Stack) SetServiceImages(images map[string]string) error {
	objs, err := s.GetObjects()
	if err != nil {
		return err
	}

	for _, obj := range objs {
		if svc, ok := obj.(*riov1.Service); ok {
			image := images[svc.Name]
			if image != "" {
				logrus.Debugf("Service %s is replaced with image %s", svc.Name, image)
				svc.Spec.Image = image
			}
		}
	}
	s.objects = objs
	return nil
}
