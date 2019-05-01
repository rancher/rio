package settings

const (
	AutoScaleStack           = "rio-autoscaler"
	BuildStackName           = "build"
	CertManagerImageType     = "CERT_MANAGER_IMAGE"
	ClusterDomainName        = "cluster-domain"
	DefaultServiceVersion    = "v0"
	GatewaySecretName        = "rio-certs"
	Grafana                  = "grafana"
	IstioGatewayDeploy       = "istio-gateway"
	IstioGateway             = "istio-gateway"
	IstioPilot               = "istio-pilot"
	IstioStackName           = "istio"
	IstioTelemetry           = "istio-telemetry"
	IstioMeshConfigKey       = "meshConfig"
	IstionConfigMapName      = "mesh"
	IstioSidecarTemplateName = "sidecarTemplate"
	ProductionIssuerName     = "letsencrypt-production-issuer"
	ProductionType           = "production"
	Prometheus               = "prometheus"
	PublicDomainType         = "RIO_PUBLICDOMAIN_CERT_TYPE"
	RDNSURL                  = "https://api.on-rio.io/v1"
	RioGateway               = "rio-gateway"
	RioWildcardType          = "RIO_WILDCARD_CERT_TYPE"
	SelfSignedIssuerName     = "selfsigned-issuer"
	SelfSignedType           = "selfsigned"
	StagingIssuerName        = "letsencrypt-staging-issuer"
	StagingType              = "staging"
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
