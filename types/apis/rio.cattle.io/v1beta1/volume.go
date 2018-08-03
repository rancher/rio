package v1beta1

import (
	"github.com/rancher/norman/types"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Volume struct {
	types.Namespaced

	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VolumeSpec   `json:"spec,omitempty"`
	Status VolumeStatus `json:"status,omitempty"`
}

type VolumeSpec struct {
	Description string `json:"description,omitempty"`
	Driver      string `json:"driver,omitempty"`
	Template    bool   `json:"template,omitempty,noupdate"`
	SizeInGB    int    `json:"sizeInGb,omitempty,required"`
	AccessMode  string `json:"accessMode,omitempty" norman:"type=enum,options=readWriteOnce|readOnlyMany|readWriteMany,default=readWriteOnce"`
	StackScoped
}

type VolumeStatus struct {
	PVCStatus  *v1.PersistentVolumeClaimStatus `json:"pvcStatus,omitempty"`
	Conditions []Condition                     `json:"conditions,omitempty"`
}
