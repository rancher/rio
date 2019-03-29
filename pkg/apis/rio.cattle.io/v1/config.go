package v1

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"github.com/rancher/wrangler/pkg/genericcondition"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigSpec   `json:"spec,omitempty"`
	Status ConfigStatus `json:"status,omitempty"`
}

type ConfigStatus struct {
	Conditions []genericcondition.GenericCondition `json:"conditions,omitempty"`
}

type ConfigSpec struct {
	Description string `json:"description,omitempty"`
	Content     string `json:"content,omitempty"`
	Encoded     bool   `json:"encoded,omitempty"`
}

type SecretMapping struct {
	Source string `json:"source,omitempty" norman:"required"`
	Target string `json:"target,omitempty"`
	Mode   string `json:"mode,omitempty"`
}

type ConfigMapping struct {
	Source string `json:"source,omitempty" norman:"required"`
	Target string `json:"target,omitempty"`
	UID    int    `json:"uid,omitempty"`
	GID    int    `json:"gid,omitempty"`
	Mode   string `json:"mode,omitempty"`
}

func (c Config) Hash() (string, error) {
	content := []byte(c.Spec.Content)
	if c.Spec.Encoded {
		bytes, err := base64.StdEncoding.DecodeString(c.Spec.Content)
		if err != nil {
			return "", err
		}
		content = bytes
	}

	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:]), nil
}
