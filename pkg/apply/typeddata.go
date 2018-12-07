package apply

import (
	"github.com/rancher/rio/pkg/settings"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/types/client/rio/v1"
	"k8s.io/api/core/v1"
)

func (d *Data) AddNamespace(namespace *v1.Namespace) {
	d.AddNamespaces(map[string]*v1.Namespace{
		namespace.Name: namespace,
	})
}

func (d *Data) AddNamespaces(namespaces map[string]*v1.Namespace) {
	d.Add("", v1.GroupName, "Namespace", namespaces)
}

func (d *Data) AddStack(namespace string, stack *riov1.Stack) {
	d.AddStacks(namespace, map[string]*riov1.Stack{
		stack.Name: stack,
	})
}

func (d *Data) AddStacks(namespace string, stacks map[string]*riov1.Stack) {
	d.Add(namespace, v1.GroupName, client.StackType, stacks)
}

func (d *Data) AddService(namespace string, service *v1.Service) {
	d.AddServices(namespace, map[string]*v1.Service{
		service.Name: service,
	})
}

func (d *Data) AddServices(namespace string, services map[string]*v1.Service) {
	d.Add(settings.IstioExternalLBNamespace, v1.GroupName, "Service", services)
}
