package v3

import (
	"github.com/rancher/wrangler/pkg/genericcondition"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type User struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	DisplayName        string     `json:"displayName,omitempty"`
	Description        string     `json:"description"`
	Username           string     `json:"username,omitempty"`
	Password           string     `json:"password,omitempty" norman:"writeOnly,noupdate"`
	MustChangePassword bool       `json:"mustChangePassword,omitempty"`
	PrincipalIDs       []string   `json:"principalIds,omitempty" norman:"type=array[reference[principal]]"`
	Me                 bool       `json:"me,omitempty"`
	Enabled            *bool      `json:"enabled,omitempty" norman:"default=true"`
	Spec               UserSpec   `json:"spec,omitempty"`
	Status             UserStatus `json:"status"`
}

type UserStatus struct {
	Conditions []genericcondition.GenericCondition `json:"conditions"`
}

type UserSpec struct{}
