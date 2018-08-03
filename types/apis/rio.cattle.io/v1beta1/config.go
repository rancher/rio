package v1beta1

import (
	"fmt"

	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"github.com/rancher/norman/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Config struct {
	types.Namespaced

	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ConfigSpec `json:"spec,omitempty"`
}

type ConfigSpec struct {
	Description string `json:"description,omitempty"`
	Content     string `json:"content,omitempty"`
	Encoded     bool   `json:"encoded,omitempty"`
	StackScoped
}

type SecretMapping struct {
	Source string `json:"source,omitempty" norman:"required"`
	Target string `json:"target,omitempty"`
	Mode   string `json:"mode,omitempty"`
}

func (s SecretMapping) MaybeString() interface{} {
	if s.Target == "/"+s.Source {
		s.Target = ""
	}

	msg := s.Source
	if s.Target != "" {
		msg += ":" + s.Target
	}

	if s.Mode != "" {
		msg = fmt.Sprintf("%s,mode=%s", msg, s.Mode)
	}

	return msg
}

type ConfigMapping struct {
	Source string `json:"source,omitempty" norman:"required"`
	Target string `json:"target,omitempty"`
	UID    int    `json:"uid,omitempty"`
	GID    int    `json:"gid,omitempty"`
	Mode   string `json:"mode,omitempty"`
}

func (c ConfigMapping) MaybeString() interface{} {
	if c.Target == "/"+c.Source {
		c.Target = ""
	}

	msg := c.Source
	if c.Target != "" {
		msg += ":" + c.Target
	}

	if c.UID > 0 {
		msg = fmt.Sprintf("%s,uid=%d", msg, c.UID)
	}

	if c.GID > 0 {
		msg = fmt.Sprintf("%s,gid=%d", msg, c.GID)
	}

	if c.Mode != "" {
		msg = fmt.Sprintf("%s,mode=%s", msg, c.Mode)
	}

	return msg
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
