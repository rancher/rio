package constants

const (
	InstallModeIngress  = "ingress"
	InstallModeSvclb    = "svclb"
	InstallModeHostport = "hostport"

	ServiceMeshModeLinkerd = "linkerd"
	ServiceMeshModeIstio   = "istio"

	ClusterIngressName = "cluster-ingress"

	L5dOverrideHeader = "l5d-dst-override"
	L5dRemoteIPHeader = "l5d-remote-ip"
	L5dServerIDHeader = "l5d-server-id"

	FeatureBuild        = "build"
	FeatureIstio        = "istio"
	FeatureGrafana      = "grafana"
	FeatureMixer        = "mixer"
	FeatureKiali        = "kiali"
	FeaturePromethues   = "prometheus"
	FeatureGateway      = "gateway"
	FeatureLetsencrypts = "letsencrypt"
	FeatureAutoscaling  = "autoscaling"

	AuthWebhookSecretName  = "auth-webhook"
	AuthWebhookServiceName = "auth-webhook"

	DevWebhookPort = ":7443"
)

var (
	ControllerImage       = "rancher/rio-controller"
	ControllerImageTag    = "dev"
	ClusterDomainName     = "cluster-domain"
	DefaultHTTPOpenPort   = "9080"
	DefaultHTTPSOpenPort  = "9443"
	InstallMode           = InstallModeSvclb
	UseIPAddress          = ""
	ServiceCidr           = ""
	DefaultServiceVersion = "v0"
	GatewaySecretName     = "rio-certs"

	IstioMeshConfigKey             = "meshConfig"
	IstionConfigMapName            = "mesh"
	IstioSidecarTemplateName       = "sidecarTemplate"
	IstioStackName                 = "istio"
	IstioTelemetry                 = "istio-telemetry"
	LetsEncryptProductionServerURL = "https://acme-v02.api.letsencrypt.org/directory"
	LetsEncryptStagingServerURL    = "https://acme-staging-v02.api.letsencrypt.org/directory"
	ProductionIssuerName           = "letsencrypt-production-issuer"
	ProductionType                 = "production"
	Prometheus                     = "prometheus"
	PublicDomainType               = "RIO_PUBLICDOMAIN_CERT_TYPE"
	RDNSURL                        = "https://api.on-rio.io/v1"
	RioGateway                     = "rio-gateway"
	RioWildcardType                = "RIO_WILDCARD_CERT_TYPE"
	SelfSignedIssuerName           = "selfsigned-issuer"
	SelfSignedType                 = "selfsigned"
	StagingIssuerName              = "letsencrypt-staging-issuer"
	StagingType                    = "staging"

	IstioVersion        = "1.2.5"
	LinkerdVersion      = "stable-2.5.0"
	LinkerdInstallImage = "rancher/linkerd-install:stable-2.5.0"

	DisableAutoscaling = false
	DisableBuild       = false
	DisableGrafana     = false
	DisableIstio       = false
	DisableKiali       = false
	DisableLinkerd     = false
	DisableLetsencrypt = false
	DisableMixer       = false
	DisablePrometheus  = false
	DisableRdns        = false

	ServiceMeshMode = "linkerd"
	GatewayName     = "gateway"

	AcmeVersion = "cm-acme"
	InstallUUID = ""

	DevMode = ""
)
