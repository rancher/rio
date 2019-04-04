package serviceset

import (
	"fmt"

	"github.com/rancher/mapper/convert"
	"github.com/rancher/mapper/convert/merge"
	"github.com/rancher/rio/cli/pkg/constants"
	"github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/pretty/schema"
)

func combineAndNormalize(name string, base map[string]interface{}, rev *riov1.Service) (*riov1.Service, error) {
	data, err := convert.EncodeToMap(rev)
	if err != nil {
		return nil, err
	}
	s := schema.Schemas.Schema(types.ServiceType)
	data = merge.UpdateMerge(s.InternalSchema, schema.Schemas, base, data, false)

	newRev := rev.DeepCopy()
	err = convert.ToObj(data, &newRev)
	if err != nil {
		return nil, err
	}

	newRev.Spec.Revision.App = name
	return newRev, nil
}

func servicesByParent(services []*riov1.Service) (Services, error) {
	result := Services{}

	for _, service := range services {
		if service.Spec.Revision.ParentService == "" {
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
	if service.Spec.Revision.Version == "" {
		service.Spec.Revision.Version = constants.DefaultServiceVersion
	}
	service.Spec.Revision.App = name
	if service.Spec.Revision.Version == constants.DefaultServiceVersion {
		service.Name = name
	} else {
		service.Name = name + "-" + service.Spec.Revision.Version
	}

	return service
}

func mergeRevisions(name string, serviceSet *ServiceSet) ([]*riov1.Service, error) {
	base, err := convert.EncodeToMap(serviceSet.Service)
	if err != nil {
		return nil, err
	}

	var newRevisions []*riov1.Service
	for _, rev := range serviceSet.Revisions {
		rev, err := combineAndNormalize(name, base, rev)
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

		name := service.Service.Spec.Revision.App
		if name == "" {
			name = service.Service.Name
		}

		service.Service = normalizeParent(name, service.Service)
		service.Revisions, err = mergeRevisions(name, service)
		if err != nil {
			return nil, err
		}

		result[name] = service
	}

	return result, nil
}
