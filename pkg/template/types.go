package template

import (
	"github.com/rancher/rio/pkg/pretty"
	"github.com/rancher/types/apis/management.cattle.io/v3"
)

type Template struct {
	Namespace       string
	Meta            pretty.TemplateMeta
	Content         []byte
	AdditionalFiles map[string][]byte
	Answers         map[string]string
	Questions       []v3.Question
}
