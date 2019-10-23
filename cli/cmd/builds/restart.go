package builds

import (
	"fmt"

	"github.com/rancher/rio/cli/pkg/clicontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Restart struct {
}

func (r *Restart) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return fmt.Errorf("at least one argument is required")
	}
	for _, arg := range ctx.CLI.Args() {
		tr, err := ctx.ByID(arg)
		if err != nil {
			return err
		}
		if err := ctx.Build.TaskRuns(tr.Namespace).Delete(tr.Name, &metav1.DeleteOptions{}); err != nil {
			return err
		}
	}

	return nil
}
