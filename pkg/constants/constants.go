package constants

const (
	AuthWebhookSecretName  = "auth-webhook"
	AuthWebhookServiceName = "auth-webhook"

	DevWebhookPort   = ":7443"
	RegistryService  = "localhost:80"
	LocalRegistry    = "localhost:5442"
	BuildkitdService = "buildkitd"
	BuildkitdImage   = "moby/buildkit:v0.6.1"

	AutoscalerServiceName = "autoscaler"

	DefaultGitCrendential    = "gitcredential"
	DefaultGitCrendentialSSH = "gitcredential-ssh"
	DefaultDockerCrendential = "dockerconfig"
	DefaultGithubCrendential = "githubtoken"
)

var (
	ControllerImage    = "rancher/rio-controller"
	ControllerImageTag = "dev"

	Prometheus = "prometheus"
	RDNSURL    = "https://api.on-rio.io/v1"

	LinkerdInstallImage = "rancher/linkerd-install:stable-2.6.0"

	DevMode = false

	DefaultStorageClass = false
	RegistryStorageSize = "20Gi"
)
