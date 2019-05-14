package rm

import (
	"strings"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	clitypes "github.com/rancher/rio/cli/pkg/types"
)

type Rm struct {
	T_Type string `desc:"delete specific type"`
}

func (r *Rm) Run(ctx *clicontext.CLIContext) error {
	types := []string{clitypes.ServiceType, clitypes.PodType, clitypes.ConfigType, clitypes.RouterType, clitypes.PublicDomainType, clitypes.ExternalServiceType, clitypes.AppType, clitypes.SecretType}
	if len(r.T_Type) > 0 {
		types = []string{r.T_Type}
	}

	return Remove(ctx, types...)
}

func Remove(ctx *clicontext.CLIContext, types ...string) error {
	for _, arg := range ctx.CLI.Args() {
		if strings.Contains(arg, ":") {
			types = []string{clitypes.ServiceType}
		} else {
			for i, t := range types {
				if t == clitypes.ServiceType {
					types = append(types[0:i], types[i+1:]...)
					break
				}
			}
		}
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
