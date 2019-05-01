package rm

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	clitypes "github.com/rancher/rio/cli/pkg/types"
)

type Rm struct {
	T_Type string `desc:"delete specific type"`
}

func (r *Rm) Run(ctx *clicontext.CLIContext) error {
	types := []string{clitypes.ServiceType, clitypes.StackType, clitypes.PodType, clitypes.ConfigType, clitypes.RouterType, clitypes.VolumeType, clitypes.ExternalServiceType}
	if len(r.T_Type) > 0 {
		types = []string{r.T_Type}
	}

	return Remove(ctx, types...)
}

func Remove(ctx *clicontext.CLIContext, types ...string) error {
	for _, arg := range ctx.CLI.Args() {
		resource, err := lookup.Lookup(ctx, arg, types...)
		if err != nil {
			return err
		}

		if err := ctx.DeleteResource(resource); err != nil {
			return err
		}
	}

	return nil
}
