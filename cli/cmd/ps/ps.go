package ps

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/urfave/cli"
)

type Ps struct {
	C_Containers bool   `desc:"print containers, not services"`
	N_Namespace  string `desc:"specify namespace"`
	System       bool   `desc:"whether to show system resources"`
}

func (p *Ps) Customize(cmd *cli.Command) {
	cmd.Flags = append(table.WriterFlags(), cmd.Flags...)
}

func (p *Ps) Run(ctx *clicontext.CLIContext) error {
	if p.C_Containers {
		return p.containers(ctx)
	}

	return p.apps(ctx)
}
