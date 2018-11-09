package publicdomain

import (
	"github.com/rancher/rio/cli/cmd/rm"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/types/client/space/v1beta1"
)

type Rm struct {
}

func (r *Rm) Run(ctx *clicontext.CLIContext) error {
	return rm.Remove(ctx, client.PublicDomainType)
}
