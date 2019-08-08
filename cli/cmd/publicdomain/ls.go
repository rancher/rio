package publicdomain

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/cli/pkg/tables"
	"github.com/rancher/rio/cli/pkg/types"
	"github.com/urfave/cli"
)

type Ls struct {
}

func (l *Ls) Customize(cmd *cli.Command) {
	cmd.Flags = append(table.WriterFlags(), cmd.Flags...)
}

func (l *Ls) Run(ctx *clicontext.CLIContext) error {
	_, err := ListPublicDomain(ctx)
	return err
}

func ListPublicDomain(ctx *clicontext.CLIContext) (bool, error) {
	publicDomains, err := ctx.List(types.PublicDomainType)
	if err != nil {
		return false, err
	}

	writer := tables.NewPublicDomain(ctx)
	return len(publicDomains) == 0, writer.Write(publicDomains)
}
