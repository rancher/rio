package kill

import (
	"fmt"

	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Kill struct {
}

func (k *Kill) Run(ctx *clicontext.CLIContext) error {
	args := ctx.CLI.Args()
	if len(args) != 1 {
		return fmt.Errorf("kill command requires exactly one argument")
	}

	pds, err := util.ListPods(ctx, args[0])
	if err != nil {
		return err
	}

	for _, pd := range pds {
		if err := ctx.Core.Pods(pd.Namespace).Delete(pd.Name, &metav1.DeleteOptions{}); err != nil {
			return err
		}
		fmt.Printf("Pod %s/%s is killed\n", pd.Namespace, pd.Name)
	}
	return nil
}
