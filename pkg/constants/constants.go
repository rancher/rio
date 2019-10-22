package constants

const (
	AuthWebhookSecretName  = "auth-webhook"
	AuthWebhookServiceName = "auth-webhook"

	DevWebhookPort   = ":7443"
	RegistryService  = "localhost:80"
	BuildkitdService = "buildkitd"
	BuildkitdImage   = "moby/buildkit:v0.6.1"

	ServiceMeshName = "linkerd"

	AutoscalerServiceName = "autoscaler"

	DefaultGitCrendential    = "gitcredential"
	DefaultDockerCrendential = "dockerconfig"
	DefaultGithubCrendential = "githubtoken"
)

var (
	ControllerImage    = "rancher/rio-controller"
	ControllerImageTag = "dev"
	UseIPAddress       = ""

	Prometheus = "prometheus"
	RDNSURL    = "https://api.on-rio.io/v1"

	LinkerdInstallImage = "rancher/linkerd-install:stable-2.6.0"

	AcmeVersion = "cm-acme"
	InstallUUID = ""

	DevMode = false

	DefaultStorageClass = false
	RegistryStorageSize = "20Gi"
)
