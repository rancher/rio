package settings

const (
	ClusterDomainName        = "cluster-domain"
	DefaultServiceVersion    = "v0"
	GatewaySecretName        = "rio-certs"
	IstioGatewayDeploy       = "istio-gateway"
	IstioGateway             = "istio-gateway"
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
	LetsEncryptStagingServerURL    = "https://acme-staging-v02.api.letsencrypt.org/directory"
	LetsEncryptProductionServerURL = "https://acme-v02.api.letsencrypt.org/directory"

	DefaultHTTPOpenPort  = "80"
	DefaultHTTPSOpenPort = "443"
)
