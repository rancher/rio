package issuers

import "github.com/rancher/rio/pkg/settings"

var IssuerTypeToName = map[string]string{
	settings.StagingType:    settings.StagingIssuerName,
	settings.ProductionType: settings.ProductionIssuerName,
	settings.SelfSignedType: settings.SelfSignedIssuerName,
}

const (
	TLSSecretName    = "rio-certs"
	RioWildcardCerts = "rio-wildcard"
)
