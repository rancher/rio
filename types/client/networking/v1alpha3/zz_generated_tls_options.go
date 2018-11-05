package client

const (
	TLSOptionsType                   = "tlsOptions"
	TLSOptionsFieldCaCertificates    = "caCertificates"
	TLSOptionsFieldHTTPSRedirect     = "httpsRedirect"
	TLSOptionsFieldMode              = "mode"
	TLSOptionsFieldPrivateKey        = "privateKey"
	TLSOptionsFieldServerCertificate = "serverCertificate"
	TLSOptionsFieldSubjectAltNames   = "subjectAltNames"
)

type TLSOptions struct {
	CaCertificates    string `json:"caCertificates,omitempty" yaml:"caCertificates,omitempty"`
	HTTPSRedirect     bool   `json:"httpsRedirect,omitempty" yaml:"httpsRedirect,omitempty"`
	Mode              string `json:"mode,omitempty" yaml:"mode,omitempty"`
	PrivateKey        string `json:"privateKey,omitempty" yaml:"privateKey,omitempty"`
	ServerCertificate string `json:"serverCertificate,omitempty" yaml:"serverCertificate,omitempty"`
	SubjectAltNames   string `json:"subjectAltNames,omitempty" yaml:"subjectAltNames,omitempty"`
}
