package builds

import (
	"fmt"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/stack"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Restart struct {
}

func (r *Restart) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return fmt.Errorf("at least one argument is required")
	}
	for _, arg := range ctx.CLI.Args() {
		namespace, name := stack.NamespaceAndName(ctx, arg)
		if err := ctx.Build.Builds(namespace).Delete(name, &metav1.DeleteOptions{}); err != nil {
			return err
		}
	}

	return nil
}
