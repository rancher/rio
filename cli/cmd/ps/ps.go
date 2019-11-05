package ps

import (
	"fmt"
	"strings"

	"github.com/rancher/rio/cli/cmd/externalservice"
	"github.com/rancher/rio/cli/cmd/route"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/urfave/cli"
)

type Ps struct {
	A_All       bool `desc:"print all resources, including router and externalservice"`
	W_Workloads bool `desc:"include apps/v1 Deployments and DaemonSets in output"`
}

func (p *Ps) Customize(cmd *cli.Command) {
	cmd.Flags = append(table.WriterFlags(), cmd.Flags...)
}

func (p *Ps) Run(ctx *clicontext.CLIContext) error {
	if p.A_All {
		return p.showAll(ctx)
	}
	return p.services(ctx)
}

func (p *Ps) showAll(ctx *clicontext.CLIContext) error {
	buffer := &strings.Builder{}
	ctx.WithWriter(buffer)

	if err := p.services(ctx); err != nil {
		return err
	}
	fmt.Fprintf(buffer, "\n")

	if err := route.ListRouters(ctx); err != nil {
		return err
	}
	fmt.Fprintf(buffer, "\n")

	if err := externalservice.ListExternalServices(ctx); err != nil {
		return err
	}
	fmt.Fprintf(buffer, "\n")

	fmt.Print(buffer.String())
	return nil
}
