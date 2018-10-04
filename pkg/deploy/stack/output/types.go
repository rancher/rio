package output

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/rio/pkg/apply"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	appsv1 "k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
	v1beta12 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Deployment struct {
	Injectors              []apply.ConfigInjector
	ClusterRoleBindings    map[string]*rbacv1.ClusterRoleBinding
	ClusterRoles           map[string]*rbacv1.ClusterRole
	ConfigMaps             map[string]*v1.ConfigMap
	DaemonSets             map[string]*appsv1.DaemonSet
	Deployments            map[string]*appsv1.Deployment
	DestinationRules       map[string]*IstioObject
	PersistentVolumeClaims map[string]*v1.PersistentVolumeClaim
	PodDisruptionBudgets   map[string]*v1beta12.PodDisruptionBudget
	RoleBindings           map[string]*rbacv1.RoleBinding
	Roles                  map[string]*rbacv1.Role
	ServiceAccounts        map[string]*v1.ServiceAccount
	Services               map[string]*v1.Service
	StatefulSets           map[string]*appsv1.StatefulSet
	VirtualServices        map[string]*IstioObject
}

func NewDeployment() *Deployment {
	return &Deployment{
		ClusterRoleBindings:    map[string]*rbacv1.ClusterRoleBinding{},
		ClusterRoles:           map[string]*rbacv1.ClusterRole{},
		ConfigMaps:             map[string]*v1.ConfigMap{},
		DaemonSets:             map[string]*appsv1.DaemonSet{},
		Deployments:            map[string]*appsv1.Deployment{},
		DestinationRules:       map[string]*IstioObject{},
		PersistentVolumeClaims: map[string]*v1.PersistentVolumeClaim{},
		PodDisruptionBudgets:   map[string]*v1beta12.PodDisruptionBudget{},
		RoleBindings:           map[string]*rbacv1.RoleBinding{},
		Roles:                  map[string]*rbacv1.Role{},
		ServiceAccounts:        map[string]*v1.ServiceAccount{},
		Services:               map[string]*v1.Service{},
		StatefulSets:           map[string]*appsv1.StatefulSet{},
		VirtualServices:        map[string]*IstioObject{},
	}
}

func (d *Deployment) Deploy(ns, groupID string) error {
	ad := apply.Data{
		GroupID:   groupID,
		Injectors: d.Injectors,
	}
	ad.Add("", rbacv1.GroupName, "ClusterRoleBinding", d.ClusterRoleBindings)
	ad.Add("", rbacv1.GroupName, "ClusterRole", d.ClusterRoles)
	ad.Add(ns, v1.GroupName, "ConfigMap", d.ConfigMaps)
	ad.Add(ns, appsv1.GroupName, "DaemonSet", d.DaemonSets)
	ad.Add(ns, appsv1.GroupName, "Deployment", d.Deployments)
	ad.Add(ns, "networking.istio.io", "DestinationRule", d.DestinationRules)
	ad.Add(ns, v1.GroupName, "PersistentVolumeClaim", d.PersistentVolumeClaims)
	ad.Add(ns, v1beta12.GroupName, "PodDisruptionBudget", d.PodDisruptionBudgets)
	ad.Add(ns, rbacv1.GroupName, "RoleBinding", d.RoleBindings)
	ad.Add(ns, rbacv1.GroupName, "Roles", d.Roles)
	ad.Add(ns, v1.GroupName, "ServiceAccount", d.ServiceAccounts)
	ad.Add(ns, v1.GroupName, "Service", d.Services)
	ad.Add(ns, appsv1.GroupName, "StatefulSet", d.StatefulSets)
	ad.Add(ns, "networking.istio.io", "VirtualService", d.VirtualServices)

	return ad.Apply()
}

type IstioObject struct {
	types.Namespaced

	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec interface{} `json:"spec,omitempty"`
}

func (i *IstioObject) DeepCopyObject() runtime.Object {
	panic("not implemented")
}

type Services map[string]*ServiceSet

func (s Services) List() []*v1beta1.Service {
	var result []*v1beta1.Service
	for _, v := range s {
		if v.Service == nil {
			continue
		}
		result = append(result, v.Service)

		for _, v := range v.Revisions {
			result = append(result, v)
		}
	}

	return result
}

type ServiceSet struct {
	Service   *v1beta1.Service
	Revisions []*v1beta1.Service
}
