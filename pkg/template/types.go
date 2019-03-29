package template

import (
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/pretty/schema"
)

type Template struct {
	Meta        schema.TemplateMeta
	Content     []byte
	Answers     map[string]string
	BuiltinVars []string
	Questions   []v1.Question
}
