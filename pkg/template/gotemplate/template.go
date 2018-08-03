package gotemplate

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/rancher/rio/pkg/template/gotemplate/funcs"
)

func Apply(contents []byte, variables map[string]string) ([]byte, error) {
	// Skip templating if contents begin with '# notemplating'
	trimmedContents := strings.TrimSpace(string(contents))
	if strings.HasPrefix(trimmedContents, "#notemplating") || strings.HasPrefix(trimmedContents, "# notemplating") {
		return contents, nil
	}

	templateFuncs := sprig.HermeticTxtFuncMap()
	templateFuncs["splitPreserveQuotes"] = funcs.SplitPreserveQuotes

	t, err := template.New("template").Funcs(templateFuncs).Parse(string(contents))
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	err = t.Execute(&buf, map[string]interface{}{
		"Values": variables,
	})
	return buf.Bytes(), err
}
