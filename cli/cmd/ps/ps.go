package ps

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/urfave/cli"
)

type Ps struct {
	C_Containers bool `desc:"print containers, not services"`
	A_All        bool `desc:"include all container, not just ones related to services"`
}

func (p *Ps) Customize(cmd *cli.Command) {
	cmd.Flags = append(table.WriterFlags(), cmd.Flags...)
}

func (p *Ps) Run(ctx *clicontext.CLIContext) error {
	if p.C_Containers || len(ctx.CLI.Args()) > 0 {
		return p.containers(ctx)
	}
	return p.services(ctx)
}
