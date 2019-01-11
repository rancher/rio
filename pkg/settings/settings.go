package settings

import (
	"github.com/rancher/rio/pkg/namespace"
)

const (
	RioSystemNamespace    = "rio-system"
	IstioGateway          = "rio-gateway"
	IstioGatewayDeploy    = "istio-gateway"
	IstioStackName        = "istio"
	IstioTelemetry        = "istio-telemetry"
	Prometheus            = "prometheus"
	AutoScaleStack        = "rio-autoscaler"
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
	settings = map[string]Setting{}
	provider Provider

	LocalStacksDir = NewSetting("local-projects-dir", "/etc/rancher/rio/projects/")
	ClusterDomain  = NewSetting("cluster-domain", "")
	IstioEnabled   = NewSetting("istio", "true")
	RDNSURL        = NewSetting("rdns-url", "https://api.lb.rancher.cloud/v1")
	RioImage       = NewSetting("rio-image", "rancher/rio")

	IstioExternalLBNamespace = namespace.StackNamespace(RioSystemNamespace, IstioStackName)
	IstioTelemetryNamespace  = namespace.StackNamespace(RioSystemNamespace, IstioTelemetry)
	IstioGatewaySelector     = map[string]string{
		"gateway": "external",
	}

	LetsEncryptStagingServerUrl    = NewSetting("letsencrypt-staging-server", "https://acme-staging-v02.api.letsencrypt.org/directory")
	LetsEncryptProductionServerUrl = NewSetting("letsencrypt-production-server", "https://acme-v02.api.letsencrypt.org/directory")
	LetsEncryptAccountEmail        = NewSetting("letsencrypt-account-email", "daishan@rancher.com")
	CertManagerImage               = NewSetting("certmanager-image", "daishan1992/cert-manager:latest")

	DefaultHTTPOpenPort  = NewSetting("default-http-port", "80")
	DefaultHTTPSOpenPort = NewSetting("default-https-port", "443")

	EnableMonitoring = NewSetting("enable-monitoring", "true")
	RouteStubImage   = NewSetting("route-stub-image", "ibuildthecloud/demo:v1")
)

type Provider interface {
	Get(name string) string
	Set(name, value string) error
	SetIfUnset(name, value string) error
	SetAll(settings map[string]Setting) error
}

type Setting struct {
	Name     string
	Default  string
	ReadOnly bool
}

func (s Setting) SetIfUnset(value string) error {
	if provider == nil {
		return s.Set(value)
	}
	return provider.SetIfUnset(s.Name, value)
}

func (s Setting) Set(value string) error {
	if provider == nil {
		s, ok := settings[s.Name]
		if ok {
			s.Default = value
			settings[s.Name] = s
		}
	} else {
		return provider.Set(s.Name, value)
	}
	return nil
}

func (s Setting) Get() string {
	if provider == nil {
		s := settings[s.Name]
		return s.Default
	}
	return provider.Get(s.Name)
}

func SetProvider(p Provider) error {
	if err := p.SetAll(settings); err != nil {
		return err
	}
	provider = p
	return nil
}

func NewSetting(name, def string) Setting {
	s := Setting{
		Name:    name,
		Default: def,
	}
	settings[s.Name] = s
	return s
}
