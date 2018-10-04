package service

import (
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/norman/types/convert/merge"
	"github.com/rancher/rio/pkg/deploy/stack/output"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1/schema"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

func combineAndNormalize(name string, base map[string]interface{}, rev *v1beta1.Service) (*v1beta1.Service, error) {
	data, err := convert.EncodeToMap(rev)
	if err != nil {
		return nil, err
	}
	s := schema.Schemas.Schema(&schema.Version, client.ServiceType)
	data = merge.UpdateMerge(s.InternalSchema, schema.Schemas, base, data, false)

	newRev := rev.DeepCopy()
	err = convert.ToObj(data, &newRev)
	if err != nil {
		return nil, err
	}

	newRev.Spec.Revision.ServiceName = name
	return newRev, nil
}

func servicesByParent(services []*v1beta1.Service) output.Services {
	result := output.Services{}

	for _, service := range services {
		if service.Spec.Revision.ParentService == "" {
			s, ok := result[service.Name]
			if !ok {
				s = &output.ServiceSet{
					Revisions: []*v1beta1.Service{},
				}
				result[service.Name] = s
			}
			s.Service = service
		} else {
			s, ok := result[service.Spec.Revision.ParentService]
			if !ok {
				s = &output.ServiceSet{
					Revisions: []*v1beta1.Service{},
				}
				result[service.Spec.Revision.ParentService] = s
			}
			s.Revisions = append(s.Revisions, service)
		}
	}

	return result
}

func normalizeParent(name string, service *v1beta1.Service) *v1beta1.Service {
	service = service.DeepCopy()
	if service.Spec.Revision.Version == "" {
		service.Spec.Revision.Version = settings.DefaultServiceVersion
	}
	service.Spec.Revision.ServiceName = name
	if service.Spec.Revision.Version == settings.DefaultServiceVersion {
		service.Name = name
	} else {
		service.Name = name + "-" + service.Spec.Revision.Version
	}

	return service
}

func mergeRevisions(name string, serviceSet *output.ServiceSet) ([]*v1beta1.Service, error) {
	base, err := convert.EncodeToMap(serviceSet.Service)
	if err != nil {
		return nil, err
	}

	var newRevisions []*v1beta1.Service
	for _, rev := range serviceSet.Revisions {
		rev, err := combineAndNormalize(name, base, rev)
		if err != nil {
			return nil, err
		}
		newRevisions = append(newRevisions, rev)
	}

	return newRevisions, nil
}

func CollectionServices(services []*v1beta1.Service) (output.Services, error) {
	var err error
	byParent := servicesByParent(services)
	result := output.Services{}

	for _, service := range byParent {
		if service.Service == nil {
			continue
		}

		name := service.Service.Spec.Revision.ServiceName
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
