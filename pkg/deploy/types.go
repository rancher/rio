package deploy

import "github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"

type StackResources struct {
	Configs  []*v1beta1.Config
	Services []*v1beta1.Service
	Volumes  []*v1beta1.Volume
	RouteSet []*v1beta1.RouteSet
}
