package types

import "k8s.io/apimachinery/pkg/runtime"

const (
	ConfigType          = "configmap"
	ServiceType         = "service"
	PodType             = "pod"
	NamespaceType       = "namespace"
	RouterType          = "router"
	ExternalServiceType = "externalservice"
	FeatureType         = "feature"
	PublicDomainType    = "publicdomain"
	SecretType          = "secret"
	BuildType           = "build"
	StackType           = "stack"
)

type Resource struct {
	LookupName, Name, Namespace, Type string
	Object                            runtime.Object
}
