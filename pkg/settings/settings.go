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
	IstioTelemetry        = "istio-telemetry"
	DefaultServiceVersion = "v0"
	StagingType           = "staging"
	ProductionType        = "production"
	SelfSignedType        = "selfsigned"
	StagingIssuerName     = "letsencrypt-staging-issuer"
	ProductionIssuerName  = "letsencrypt-production-issuer"
	SelfSignedIssuerName  = "selfsigned-issuer"
	RioWildcardType       = "RIO_WILDCARD_CERT_TYPE"
	PublicDomainType      = "RIO_PUBLICDOMAIN_CERT_TYPE"
	CertManagerImageType  = "CERT_MANAGER_IMAGE"
	IstionConfigMapName   = "mesh"
	IstionConfigMapKey    = "content"
)

var (
	LocalStacksDir = settings.NewSetting("local-projects-dir", "/etc/rancher/rio/projects/")
	ClusterDomain  = settings.NewSetting("cluster-domain", "")
	IstioEnabled   = settings.NewSetting("istio", "true")
	RDNSURL        = settings.NewSetting("rdns-url", "https://api.lb.rancher.cloud/v1")
	RioImage       = settings.NewSetting("rio-image", "rancher/rio")

	IstioExternalLBNamespace = namespace.StackNamespace(RioSystemNamespace, IstioStackName)
	IstioTelemetryNamespace  = namespace.StackNamespace(RioSystemNamespace, IstioTelemetry)
	IstioGatewaySelector     = map[string]string{
		"gateway": "external",
	}

	LetsEncryptStagingServerUrl    = settings.NewSetting("letsencrypt-staging-server", "https://acme-staging-v02.api.letsencrypt.org/directory")
	LetsEncryptProductionServerUrl = settings.NewSetting("letsencrypt-production-server", "https://acme-v02.api.letsencrypt.org/directory")
	LetsEncryptAccountEmail        = settings.NewSetting("letsencrypt-account-email", "daishan@rancher.com")
	CertManagerImage               = settings.NewSetting("certmanager-image", "daishan1992/cert-manager:latest")

	DefaultHTTPOpenPort  = settings.NewSetting("default-http-port", "80")
	DefaultHTTPSOpenPort = settings.NewSetting("default-https-port", "443")

	EnableMonitoring = settings.NewSetting("enable-monitoring", "true")
	RouteStubImage   = settings.NewSetting("route-stub-image", "ibuildthecloud/demo:v1")
)
