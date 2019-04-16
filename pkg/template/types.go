package template

import (
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

type Template struct {
	Content     []byte
	BuiltinVars []string
}

type AnswerCallback func(key string, questions []v1.Question) (string, error)
