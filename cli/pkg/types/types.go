package types

import "k8s.io/apimachinery/pkg/runtime"

const (
	ConfigType          = "config"
	AppType             = "app"
	ServiceType         = "service"
	PodType             = "pod"
	NamespaceType       = "namespace"
	RouterType          = "router"
	ExternalServiceType = "externalservice"
	FeatureType         = "feature"
	PublicDomainType    = "publicdomain"
	SecretType          = "secret"
	BuildType           = "build"
)

type Resource struct {
	LookupName, Name, Namespace, Type string
	Object                            runtime.Object
}
