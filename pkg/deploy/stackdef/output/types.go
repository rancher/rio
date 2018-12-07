package output

import (
	"github.com/rancher/rio/pkg/apply"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apimachinery/pkg/runtime"
)

type Deployment struct {
	Namespace  string
	Namespaces map[string]*v1.Namespace
	Services   map[string]*riov1.Service
	Configs    map[string]*riov1.Config
	Volumes    map[string]*riov1.Volume
	Routes     map[string]*riov1.RouteSet
	CRDs       map[string]runtime.Object
	K8sObjects map[string]map[string]runtime.Object
}

func NewDeployment() *Deployment {
	return &Deployment{
		Namespaces: map[string]*v1.Namespace{},
		Services:   map[string]*riov1.Service{},
		Configs:    map[string]*riov1.Config{},
		Volumes:    map[string]*riov1.Volume{},
		Routes:     map[string]*riov1.RouteSet{},
		CRDs:       map[string]runtime.Object{},
		K8sObjects: map[string]map[string]runtime.Object{},
	}
}

func (d *Deployment) Deploy(groupID string) error {
	ad := apply.Data{
		GroupID: groupID,
	}

	ad.Add("", apiextensions.GroupName, "CustomResourceDefinition", d.CRDs)
	for ns, objs := range d.K8sObjects {
		ad.Add(ns, "", "", objs)
	}
	ad.Add("", v1.GroupName, "Namespace", d.Namespaces)
	ad.Add(d.Namespace, riov1.GroupName, "Service", d.Services)
	ad.Add(d.Namespace, riov1.GroupName, "Config", d.Configs)
	ad.Add(d.Namespace, riov1.GroupName, "Volume", d.Volumes)
	ad.Add(d.Namespace, riov1.GroupName, "RouteSet", d.Routes)

	return ad.Apply()
}
