package settings

import (
	"github.com/eggsampler/acme"
	"github.com/rancher/rancher/pkg/settings"
	"github.com/rancher/rio/pkg/namespace"
)

const (
	RioSystemNamespace    = "rio-system"
	IstioExternalLB       = "rio-lb"
	IstioGateway          = "rio-gateway"
	IstioStackName        = "istio"
	DefaultServiceVersion = "v0"
	CerManagerIssuerName  = "letsencrypt-issuer"
)

var (
	ClusterDomain = settings.NewSetting("cluster-domain", "")
	IstioEnabled  = settings.NewSetting("istio", "true")
	RDNSURL       = settings.NewSetting("rdns-url", "https://api.lb.rancher.cloud/v1")
	RioImage      = settings.NewSetting("rio-image", "rancher/rio")

	IstioExternalLBNamespace = namespace.StackNamespace(RioSystemNamespace, IstioStackName)
	IstioGatewaySelector     = map[string]string{
		"gateway": "external",
	}

	LetsEncryptServerUrl    = settings.NewSetting("letsencrypt-server", acme.LetsEncryptStaging)
	LetsEncryptAccountEmail = settings.NewSetting("letsencrypt-account-email", "daishan@rancher.com")
)
