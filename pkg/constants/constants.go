package constants

const (
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
	UseIPAddress          = ""
	DefaultServiceVersion = "v0"
	ServiceCidr           = ""

	Prometheus = "prometheus"
	RDNSURL    = "https://api.on-rio.io/v1"

	IstioVersion        = "1.2.5"
	LinkerdInstallImage = "rancher/linkerd-install:stable-2.5.0"

	GatewayName = "gateway"

	AcmeVersion = "cm-acme"
	InstallUUID = ""

	DevMode = ""
)
