package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PublicDomain struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PublicDomainSpec `json:"spec,inline"`
}

type PublicDomainSpec struct {
	RouteSetName   string `json:"routeSetName,omitempty"`
	ServiceName    string `json:"serviceName,omitempty"`
	StackName      string `json:"stackName,omitempty"`
	SpaceName      string `json:"spaceName,omitempty"`
	DomainName     string `json:"domainName,omitempty"`
	RequestTLSCert bool   `json:"requestTlsCert,omitempty"`
}
