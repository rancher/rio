package settings

const (
	AutoScaleStack                 = "rio-autoscaler"
	BuildStackName                 = "build"
	CertManagerImage               = "daishan1992/cert-manager:latest"
	CertManagerImageType           = "CERT_MANAGER_IMAGE"
	ClusterDomainName              = "cluster-domain"
	DefaultHTTPOpenPort            = "80"
	DefaultHTTPSOpenPort           = "443"
	DefaultServiceVersion          = "v0"
	Grafana                        = "grafana"
	IstioGatway                    = "istio-gateway"
	IstionConfigMapKey             = "content"
	IstioStackName                 = "istio"
	IstioTelemetry                 = "istio-telemetry"
	LetsEncryptAccountEmail        = "daishan@rancher.com"
	LetsEncryptProductionServerUrl = "https://acme-v01.api.letsencrypt.org/directory"
	LetsEncryptStagingServerUrl    = "https://acme-staging-v02.api.letsencrypt.org/directory"
	MeshConfigMapName              = "mesh"
	ProductionIssuerName           = "letsencrypt-production-issuer"
	ProductionType                 = "production"
	Prometheus                     = "prometheus"
	PublicDomainType               = "RIO_PUBLICDOMAIN_CERT_TYPE"
	RDNSURL                        = "https://api.lb.rancher.cloud/v1"
	RioGateway                     = "rio-gateway"
	RioWildcardType                = "RIO_WILDCARD_CERT_TYPE"
	SelfSignedIssuerName           = "selfsigned-issuer"
	SelfSignedType                 = "selfsigned"
	StagingIssuerName              = "letsencrypt-staging-issuer"
	StagingType                    = "staging"
)

var (
	ClusterDomain = ""
)
