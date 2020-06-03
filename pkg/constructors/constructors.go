package constructors

import (
	splitv1alpha1 "github.com/deislabs/smi-sdk-go/pkg/apis/split/v1alpha1"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
)

func NewNamespace(name string, obj v1.Namespace) *v1.Namespace {
	obj.APIVersion = "v1"
	obj.Kind = "SystemNamespace"
	obj.Name = name
	return &obj
}

func NewSecret(namespace, name string, obj v1.Secret) *v1.Secret {
	obj.APIVersion = "v1"
	obj.Kind = "Secret"
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

func NewServiceAccount(namespace, name string, obj v1.ServiceAccount) *v1.ServiceAccount {
	obj.APIVersion = "v1"
	obj.Kind = "ServiceAccount"
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

func NewConfigMap(namespace, name string, obj v1.ConfigMap) *v1.ConfigMap {
	obj.APIVersion = "v1"
	obj.Kind = "ConfigMap"
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

func NewDeployment(namespace, name string, obj appsv1.Deployment) *appsv1.Deployment {
	obj.APIVersion = "apps/v1"
	obj.Kind = "Deployment"
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

func NewStatefulSet(namespace, name string, obj appsv1.StatefulSet) *appsv1.StatefulSet {
	obj.APIVersion = "apps/v1"
	obj.Kind = "StatefulSet"
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

func NewDaemonset(namespace, name string, obj appsv1.DaemonSet) *appsv1.DaemonSet {
	obj.APIVersion = "apps/v1"
	obj.Kind = "DaemonSet"
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

func NewService(namespace, name string, obj v1.Service) *v1.Service {
	obj.APIVersion = "v1"
	obj.Kind = "Service"
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

func NewPersistentVolumeClaim(namespace, name string, obj v1.PersistentVolumeClaim) *v1.PersistentVolumeClaim {
	obj.APIVersion = "v1"
	obj.Kind = "PersistentVolumeClaim"
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

func NewIngress(namespace, name string, obj extensionsv1beta1.Ingress) *extensionsv1beta1.Ingress {
	obj.APIVersion = "extensions/v1beta1"
	obj.Kind = "Ingress"
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

func NewTrafficSplit(namespace, name string, obj splitv1alpha1.TrafficSplit) *splitv1alpha1.TrafficSplit {
	obj.APIVersion = "split.smi-spec.io/v1alpha1"
	obj.Kind = "TrafficSplit"
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

func NewEndpoints(namespace, name string, obj v1.Endpoints) *v1.Endpoints {
	obj.APIVersion = "v1"
	obj.Kind = "Endpoints"
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

func NewTaskRun(namespace, name string, obj tektonv1beta1.TaskRun) *tektonv1beta1.TaskRun {
	obj.APIVersion = "tekton.dev/v1alpha1"
	obj.Kind = "TaskRun"
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}
