package client

const (
	PublicDomainSpecType                     = "publicDomainSpec"
	PublicDomainSpecFieldDomainName          = "domainName"
	PublicDomainSpecFieldTargetName          = "targetName"
	PublicDomainSpecFieldTargetStackName     = "targetStackName"
	PublicDomainSpecFieldTargetWorkspaceName = "targetWorkspaceName"
)

type PublicDomainSpec struct {
	DomainName          string `json:"domainName,omitempty" yaml:"domainName,omitempty"`
	TargetName          string `json:"targetName,omitempty" yaml:"targetName,omitempty"`
	TargetStackName     string `json:"targetStackName,omitempty" yaml:"targetStackName,omitempty"`
	TargetWorkspaceName string `json:"targetWorkspaceName,omitempty" yaml:"targetWorkspaceName,omitempty"`
}
