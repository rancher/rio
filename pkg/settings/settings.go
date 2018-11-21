package settings

import (
	"github.com/rancher/rancher/pkg/settings"
	"github.com/rancher/rio/pkg/namespace"
)

const (
	RioSystemNamespace    = "rio-system"
	IstioGateway          = "rio-gateway"
	IstioGatewayDeploy    = "istio-gateway"
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

	LetsEncryptServerUrl    = settings.NewSetting("letsencrypt-server", "https://acme-staging-v02.api.letsencrypt.org/directory")
	LetsEncryptAccountEmail = settings.NewSetting("letsencrypt-account-email", "daishan@rancher.com")
	CertManagerImage        = settings.NewSetting("certmanager-image", "daishan1992/cert-manager:latest")

	DefaultHTTPOpenPort  = settings.NewSetting("default-http-port", "80")
	DefaultHTTPSOpenPort = settings.NewSetting("default-https-port", "443")
)
