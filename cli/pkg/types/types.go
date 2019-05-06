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

	ConfigTypeFull          = "configmaps"
	AppTypeFull             = "apps.rio.cattle.io"
	ServiceTypeFull         = "services.rio.cattle.io"
	RouterTypeFull          = "routers.rio.cattle.io"
	ExternalServiceTypeFull = "externalservices.rio.cattle.io"
	PublicDomainTypeFull    = "publicdomains.rio.cattle.io"
)

type Resource struct {
	LookupName, Name, Namespace, Type, FullType string
	Object                                      runtime.Object
}
