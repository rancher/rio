package input

import (
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
)

type Stack struct {
	Namespace string
	Space     string
	Stack     *v1beta1.Stack
	Configs   []*v1beta1.Config
	Services  []*v1beta1.Service
	Volumes   []*v1beta1.Volume
	RouteSet  []*v1beta1.RouteSet
}
