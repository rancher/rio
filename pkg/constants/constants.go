package constants

const (
	AuthWebhookSecretName  = "auth-webhook"
	AuthWebhookServiceName = "auth-webhook"

	DevWebhookPort   = ":7443"
	RegistryService  = "localhost:80"
	BuildkitdService = "buildkitd"

	ServiceMeshName = "linkerd"

	AutoscalerServiceName = "autoscaler"
)

var (
	ControllerImage    = "rancher/rio-controller"
	ControllerImageTag = "dev"
	UseIPAddress       = ""

	Prometheus = "prometheus"
	RDNSURL    = "https://api.on-rio.io/v1"

	LinkerdInstallImage = "rancher/linkerd-install:stable-2.5.0"

	AcmeVersion = "cm-acme"
	InstallUUID = ""

	DevMode = false

	DefaultStorageClass = false
	RegistryStorageSize = "20Gi"
)
