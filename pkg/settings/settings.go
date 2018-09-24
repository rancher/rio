package settings

import (
	"github.com/rancher/rancher/pkg/settings"
	"github.com/rancher/rio/pkg/namespace"
	"github.com/rancher/rio/version"
)

const (
	RioSystemNamespace  = "rio-system"
	RioDefaultNamespace = "rio-default"
	IstioExternalLB     = "rio-lb"
	IstioGateway        = "rio-gateway"
	IstioStackName      = "istio"
	IstioProxyStackName = "istio-proxy"
)

var (
	ClusterDomain = settings.NewSetting("cluster-domain", "")
	IstioEnabled  = settings.NewSetting("istio", "true")
	RDNSURL       = settings.NewSetting("rdns-url", "https://api.lb.rancher.cloud/v1")
	RioImage      = settings.NewSetting("rio-image", "rancher/rio")

	IstioExternalLBNamespace = namespace.StackNamespace(RioSystemNamespace, IstioProxyStackName)
	IstioGatewaySelector     = map[string]string{
		"gateway": "external",
	}
)

func RioFullImage() string {
	return RioImage.Get() + ":" + version.Version
}
