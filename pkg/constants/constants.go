package constants

const (
	AuthWebhookSecretName  = "rio-api-validator"
	AuthWebhookServiceName = "rio-api-validator"

	IstioRioGateway      = "rio-gateway"
	IstioSystemNamespace = "istio-system"

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

	StackLabel     = "gitwatcher.rio.cattle.io/stack"
	ServiceLabel   = "gitwatcher.rio.cattle.io/service"
	ContainerLabel = "gitwatcher.rio.cattle.io/container"

	GitCommitLabel = "gitwatcher.rio.cattle.io/git-commit"
	LogTokenLabel  = "gitwatcher.rio.cattle.io/log-token"

	LetsEncryptAccountSecretName = "letsencrypt-account"

	AcmeSolverServicName = "rio-acme-solver"
)

var (
	ControllerImage    = "rancher/rio-controller"
	ControllerImageTag = "dev"

	Prometheus = "prometheus"
	RDNSURL    = "https://api.on-rio.io/v1"

	LinkerdInstallImage = "rancher/linkerd-install:stable-2.9.0"

	RegistryStorageSize = "20Gi"
)
