package types

import "k8s.io/apimachinery/pkg/runtime"

const (
	ConfigType          = "config"
	VolumeType          = "volume"
	AppType             = "app"
	ServiceType         = "service"
	PodType             = "pod"
	StackType           = "stack"
	RouterType          = "router"
	ExternalServiceType = "externalservice"
	FeatureType         = "feature"
	PublicDomainType    = "publicdomain"
)

type Resource struct {
	LookupName, Name, Namespace, Type string
	Object                            runtime.Object
}
