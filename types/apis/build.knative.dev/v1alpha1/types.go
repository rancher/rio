package v1alpha1

import (
	buildapis "github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/rancher/norman/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Build struct {
	types.Namespaced
	types.StatusSubResourced

	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   buildapis.BuildSpec   `json:"spec,omitempty"`
	Status buildapis.BuildStatus `json:"status,omitempty"`
}
