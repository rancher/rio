package rm

import (
	"errors"
	"strings"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
)

type Rm struct {
	T_Type string `desc:"delete specific type. Available types: [config,service,router,externalservice,publicdomain,app,secret,build,stack]"`
}

func (r *Rm) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return errors.New("at least one argument is needed")
	}
	types := []string{clitypes.ServiceType, clitypes.ConfigType, clitypes.RouterType, clitypes.PublicDomainType, clitypes.ExternalServiceType, clitypes.AppType, clitypes.SecretType, clitypes.BuildType, clitypes.StackType}
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

		if err := ctx.DeleteResource(resource); err != nil && !kerrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}
