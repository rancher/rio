package deploy

import "github.com/rancher/rio/types/apis/rio.cattle.io/v1"

type StackResources struct {
	Configs  []*v1.Config
	Services []*v1.Service
	Volumes  []*v1.Volume
	RouteSet []*v1.RouteSet
}
