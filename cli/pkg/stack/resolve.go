package stack

import (
	"strings"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/wrangler/pkg/kv"
)

func NamespaceAndName(c *clicontext.CLIContext, in string) (string, string) {
	namespace, name := kv.Split(in, "/")
	if namespace != "" && name == "" {
		if !strings.HasSuffix(in, "/") {
			name = namespace
			namespace = ""
		}
	}
	if namespace == "" {
		namespace = c.GetSetNamespace()
	}
	if namespace == "" {
		namespace = "default"
	}
	return namespace, name
}
