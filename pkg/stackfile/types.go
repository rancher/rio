package stackfile

import (
	"github.com/rancher/rio/pkg/template"
)

type StackFile struct {
	name            string
	Template        template.Template
	AdditionalFiles map[string][]byte
}
