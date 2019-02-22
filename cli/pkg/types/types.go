package types

const (
	ConfigType          = "config"
	VolumeType          = "volume"
	ServiceType         = "service"
	PodType             = "pod"
	StackType           = "stack"
	RouteSetType        = "routeset"
	ExternalServiceType = "externalservice"
	FeatureType         = "feature"
	NodeType            = "node"
	PublicDomainType    = "publicdomain"
)

type Resource struct {
	Name, Namespace, Type string
}
