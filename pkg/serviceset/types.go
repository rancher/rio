package serviceset

import (
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

type Services map[string]*ServiceSet

func (s Services) List() []*riov1.Service {
	var result []*riov1.Service
	for _, v := range s {
		result = append(result, v.Revisions...)
	}
	return result
}

type ServiceSet struct {
	Revisions []*riov1.Service
}
