package deploy

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

func services(objects []runtime.Object, configs map[string]*v1beta1.Config, stack *StackResources, namespace string) ([]runtime.Object, error) {
	var err error

	volumeDefs := volumeMap(stack.Volumes)

	for _, s := range stack.Services {
		objects, err = service(objects, s.Name, namespace, s, volumeDefs, configs)
		if err != nil {
			return objects, errors.Wrapf(err, "failed to construct service for %s/%s", s.Namespace, s.Name)
		}
	}

	return objects, nil
}

func MergeRevisionToService(service *v1beta1.Service, revision string) (*v1beta1.ServiceUnversionedSpec, error) {
	// TODO: do better merging
	newRevision := service.Spec.ServiceUnversionedSpec.DeepCopy()
	serviceRevision, ok := service.Spec.Revisions[revision]
	if !ok {
		return nil, fmt.Errorf("failed to find revision for %s", revision)
	}

	err := convert.ToObj(&serviceRevision.Spec, newRevision)
	return newRevision, err
}

func service(objects []runtime.Object, name, namespace string, service *v1beta1.Service, volumeDefs map[string]*v1beta1.Volume, configs map[string]*v1beta1.Config) ([]runtime.Object, error) {
	objects, err := addService(objects, "latest", name, namespace, &service.Spec.ServiceUnversionedSpec, volumeDefs, configs)

	for revision := range service.Spec.Revisions {
		newRevision, err := MergeRevisionToService(service, revision)
		if err != nil {
			return nil, err
		}

		objects, err = addService(objects, revision, name, namespace, newRevision, volumeDefs, configs)
		if err != nil {
			return nil, err
		}
	}

	return objects, err
}

func addService(objects []runtime.Object, revision, serviceName, namespace string, service *v1beta1.ServiceUnversionedSpec,
	volumeDefs map[string]*v1beta1.Volume, configs map[string]*v1beta1.Config) ([]runtime.Object, error) {
	var (
		err error
	)

	labels := map[string]string{
		"rio.cattle.io": "true",
		"app":           serviceName,
		"rio.cattle.io/service":   serviceName,
		"rio.cattle.io/namespace": namespace,
		"rio.cattle.io/revision":  revision,
	}

	name := fmt.Sprintf("%s-%s", serviceName, revision)
	if revision == "latest" {
		name = serviceName
	}

	usedTemplates, podSpec := podSpec(serviceName, service.Labels, service, volumeDefs)

	objects, useServiceAccount, err := addRoles(objects, name, namespace, labels, service)
	if err != nil {
		return objects, err
	}
	if useServiceAccount {
		objects, err = addGlobalRoles(objects, name, namespace, labels, service)
		if err != nil {
			return objects, err
		}
		podSpec.ServiceAccountName = name
		podSpec.AutomountServiceAccountToken = nil
	}

	podTemplateSpec := podTemplateSpec(labels, service.Metadata, service.Labels, configs, podSpec)

	if service.Global {
		if len(usedTemplates) > 0 {
			return nil, fmt.Errorf("globally scheduling services can not use volume templates")
		}
		objects, err = daemonset(objects, labels, name, namespace, service, podTemplateSpec)
	} else if isDeployment(service, usedTemplates) {
		objects = deployment(objects, labels, name, namespace, service, podTemplateSpec)
	} else {
		objects, err = statefulset(objects, labels, name, namespace, service, usedTemplates, podTemplateSpec)
	}
	if err != nil {
		return objects, err
	}
	objects = serviceSelector(objects, name, namespace, service, labels)
	objects = nodePorts(objects, name+"-ports", namespace, service, labels)
	objects = pdbs(objects, name, namespace, labels, service)

	return objects, nil
}
