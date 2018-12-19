package serviceset

import (
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

type Services map[string]*ServiceSet

func (s Services) List() []*riov1.Service {
	var result []*riov1.Service
	for _, v := range s {
		result = append(result, v.List()...)
	}
	return result
}

type ServiceSet struct {
	Service   *riov1.Service
	Revisions []*riov1.Service
}

func (s ServiceSet) List() []*riov1.Service {
	var result []*riov1.Service

	if s.Service == nil {
		return nil
	}

	result = append(result, s.Service)

	for _, v := range s.Revisions {
		result = append(result, v)
	}

	return result
}
