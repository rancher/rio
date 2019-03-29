package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Volume struct {
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
}

type VolumeStatus struct {
	PVCStatus  *v1.PersistentVolumeClaimStatus `json:"pvcStatus,omitempty"`
	Conditions []Condition                     `json:"conditions,omitempty"`
}
