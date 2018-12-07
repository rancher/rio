package client

const (
	PublicDomainSpecType                   = "publicDomainSpec"
	PublicDomainSpecFieldDomainName        = "domainName"
	PublicDomainSpecFieldTargetName        = "targetName"
	PublicDomainSpecFieldTargetProjectName = "targetProjectName"
	PublicDomainSpecFieldTargetStackName   = "targetStackName"
)

type PublicDomainSpec struct {
	DomainName        string `json:"domainName,omitempty" yaml:"domainName,omitempty"`
	TargetName        string `json:"targetName,omitempty" yaml:"targetName,omitempty"`
	TargetProjectName string `json:"targetProjectName,omitempty" yaml:"targetProjectName,omitempty"`
	TargetStackName   string `json:"targetStackName,omitempty" yaml:"targetStackName,omitempty"`
}
