package v1alpha3

import (
	"github.com/rancher/norman/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Gateway struct {
	types.Namespaced

	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GatewaySpec `json:"spec"`
}

type VirtualService struct {
	types.Namespaced

	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

type DestinationRule struct {
	types.Namespaced

	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

type GatewaySpec struct {
	Servers  []*Server         `json:"servers,omitempty"`
	Selector map[string]string `json:"selector,omitempty"`
}

type Server struct {
	Port  *Port       `json:"port,omitempty"`
	Hosts []string    `json:"hosts,omitempty"`
	TLS   *TLSOptions `json:"tls,omitempty"`
}

type TLSOptions struct {
	HTTPSRedirect     bool   `json:"httpsRedirect,omitempty"`
	Mode              string `json:"mode,omitempty"`
	ServerCertificate string `json:"serverCertificate,omitempty"`
	PrivateKey        string `json:"privateKey,omitempty"`
	CaCertificates    string `json:"caCertificates,omitempty"`
	SubjectAltNames   string `json:"subjectAltNames,omitempty"`
}

type Port struct {
	Number   uint32 `json:"number,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	Name     string `json:"name,omitempty"`
}
