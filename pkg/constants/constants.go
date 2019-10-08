package constants

const (
	ClusterIngressName = "cluster-ingress"
	ClusterDomainName  = "cluster-domain"
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
)
