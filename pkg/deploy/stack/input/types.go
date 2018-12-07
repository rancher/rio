package input

import "github.com/rancher/rio/types/apis/rio.cattle.io/v1"

type Stack struct {
	Namespace        string
	Project          string
	Stack            *v1.Stack
	Configs          []*v1.Config
	Services         []*v1.Service
	Volumes          []*v1.Volume
	RouteSet         []*v1.RouteSet
	ExternalServices []*v1.ExternalService
}
