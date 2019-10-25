package stack

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/riofile"
	"github.com/rancher/rio/pkg/services"
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
	rf, err := riofile.Parse(s.contents, template.AnswersFromMap(s.answers))
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

	rf, err := riofile.Parse(s.contents, template.AnswersFromMap(s.answers))
	if err != nil {
		return nil, err
	}
	return rf.Objects(), nil
}

type ContainerBuildKey struct {
	Service   string
	Container string
}

func (c ContainerBuildKey) String() string {
	return fmt.Sprintf("%s-%s", c.Service, c.Container)
}

func (s *Stack) GetImageBuilds() (map[ContainerBuildKey]riov1.ImageBuildSpec, error) {
	objs, err := s.GetObjects()
	if err != nil {
		return nil, err
	}

	buildConfig := make(map[ContainerBuildKey]riov1.ImageBuildSpec)
	for _, obj := range objs {
		if svc, ok := obj.(*riov1.Service); ok {
			for _, container := range services.ToNamedContainers(svc) {
				if container.ImageBuild != nil {
					if svc.Spec.ImageBuild.Repo != "" {
						continue
					}
					buildConfig[ContainerBuildKey{
						Service:   svc.Name,
						Container: container.Name,
					}] = *container.ImageBuild
				} else {
					// check if the image points toward a local path
					if strings.HasPrefix(svc.Spec.Image, "./") || strings.HasPrefix(svc.Spec.Image, "/") {
						fileInfo, err := os.Stat(svc.Spec.Image)
						if err != nil {
							return nil, errors.Wrap(err, "error parsing image field")
						}
						if fileInfo.IsDir() {
							buildConfig[ContainerBuildKey{
								Service:   svc.Name,
								Container: container.Name,
							}] = riov1.ImageBuildSpec{Context: svc.Spec.Image}
							svc.Spec.Image = ""
						} else {
							return nil, errors.Wrap(err, "image field is not a directory")
						}
					}
				}
			}
		}
	}
	return buildConfig, nil
}

func (s *Stack) SetServiceImages(images map[ContainerBuildKey]string) error {
	objs, err := s.GetObjects()
	if err != nil {
		return err
	}

	for _, obj := range objs {
		if svc, ok := obj.(*riov1.Service); ok {
			var keys []ContainerBuildKey
			for _, con := range svc.Spec.Sidecars {
				keys = append(keys, ContainerBuildKey{
					Service:   svc.Name,
					Container: con.Name,
				})
			}
			keys = append(keys, ContainerBuildKey{
				Service:   svc.Name,
				Container: svc.Name,
			})

			for _, key := range keys {
				image := images[key]
				if image != "" {
					logrus.Debugf("Service %s is replaced with image %s", svc.Name, image)
					svc.Spec.Image = image
				}
			}
		}
	}
	s.objects = objs
	return nil
}
