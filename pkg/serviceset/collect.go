package serviceset

import (
	"encoding/json"
	"fmt"

	"github.com/rancher/mapper/convert"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

func combineAndNormalize(base map[string]interface{}, rev *riov1.Service) (*riov1.Service, error) {
	baseBytes, err := json.Marshal(base)
	if err != nil {
		return nil, err
	}

	overlay, err := json.Marshal(rev)
	if err != nil {
		return nil, err
	}

	newBytes, err := strategicpatch.StrategicMergePatch(baseBytes, overlay, &riov1.Service{})
	if err != nil {
		return nil, err
	}

	result := &riov1.Service{}
	return result, json.Unmarshal(newBytes, result)
}

func servicesByParent(services []*riov1.Service) (Services, error) {
	result := Services{}

	for _, service := range services {
		if service.Spec.App == "" {
			s, ok := result[service.Name]
			if !ok {
				s = &ServiceSet{
					Revisions: []*riov1.Service{},
				}
				result[service.Name] = s
			}
			if s.Service != nil {
				return result, fmt.Errorf("two root services exist for %s", service.Name)
			}
			s.Service = service
		} else {
			s, ok := result[service.Spec.Revision.ParentService]
			if !ok {
				s = &ServiceSet{
					Revisions: []*riov1.Service{},
				}
				result[service.Spec.Revision.ParentService] = s
			}
			s.Revisions = append(s.Revisions, service)
		}
	}

	return result, nil
}

func normalizeParent(name string, service *riov1.Service) *riov1.Service {
	service = service.DeepCopy()
	if service.Spec.Version == "" {
		service.Spec.Version = constants.DefaultServiceVersion
	}
	service.Spec.App = name
	if service.Spec.Version == constants.DefaultServiceVersion {
		service.Name = name
	} else {
		service.Name = name + "-" + service.Spec.Version
	}

	return service
}

func mergeRevisions(serviceSet *ServiceSet) ([]*riov1.Service, error) {
	base, err := convert.EncodeToMap(serviceSet.Service)
	if err != nil {
		return nil, err
	}

	var newRevisions []*riov1.Service
	for _, rev := range serviceSet.Revisions {
		rev, err := combineAndNormalize(base, rev)
		if err != nil {
			return nil, err
		}
		newRevisions = append(newRevisions, rev)
	}

	return newRevisions, nil
}

func CollectionServices(services []*riov1.Service) (Services, error) {
	var err error
	result := Services{}
	byParent, err := servicesByParent(services)
	if err != nil {
		return result, err
	}

	for _, service := range byParent {
		if service.Service == nil {
			continue
		}

		name := service.Service.Spec.App
		if name == "" {
			name = service.Service.Name
		}

		service.Service = normalizeParent(name, service.Service)
		service.Revisions, err = mergeRevisions(service)
		if err != nil {
			return nil, err
		}

		result[name] = service
	}

	return result, nil
}
