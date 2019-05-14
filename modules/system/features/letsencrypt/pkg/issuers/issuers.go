package issuers

import "github.com/rancher/rio/pkg/constants"

var IssuerTypeToName = map[string]string{
	constants.StagingType:    constants.StagingIssuerName,
	constants.ProductionType: constants.ProductionIssuerName,
	constants.SelfSignedType: constants.SelfSignedIssuerName,
}

const (
	RioWildcardCerts = "rio-wildcard"
)
