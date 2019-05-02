package ps

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/urfave/cli"
)

type Ps struct {
	C_Containers bool `desc:"print containers, not services"`
	R_Revisions  bool `desc:"print revisions"`
	A_All        bool `desc:"include all container, not just ones related to services"`
	System       bool `desc:"whether to show system resources"`
}

func (p *Ps) Customize(cmd *cli.Command) {
	cmd.Flags = append(table.WriterFlags(), cmd.Flags...)
}

func (p *Ps) Run(ctx *clicontext.CLIContext) error {
	if p.C_Containers {
		return p.containers(ctx)
	}

	if p.R_Revisions {
		return p.revisions(ctx)
	}

	return p.apps(ctx)
}
