package settings

const (
	RioSystemNamespace    = "rio-system"
	IstioGateway          = "rio-gateway"
	IstioGatewayDeploy    = "istio-gateway"
	IstioStackName        = "istio"
	GatewaySecretName     = "rio-certs"
	RioGateway            = "rio-gateway"
	IstioTelemetry        = "istio-telemetry"
	Prometheus            = "prometheus"
	Grafana               = "grafana"
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
	IstioExternalLBNamespace = IstioStackName
	IstioTelemetryNamespace  = IstioTelemetry
	PrometheusNamespace      = Prometheus
	GrafanaNamespace         = Grafana
	IstioGatewaySelector     = map[string]string{
		"gateway": "external",
	}

	LetsEncryptStagingServerUrl    = "https://acme-staging-v02.api.letsencrypt.org/directory"
	LetsEncryptProductionServerUrl = "https://acme-v02.api.letsencrypt.org/directory"
	LetsEncryptAccountEmail        = "daishan@rancher.com"
	CertManagerImage               = "daishan1992/cert-manager:latest"

	DefaultHTTPOpenPort  = "80"
	DefaultHTTPSOpenPort = "443"
)
