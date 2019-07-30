package ps

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/rio/cli/cmd/externalservice"
	"github.com/rancher/rio/cli/cmd/revision"
	"github.com/rancher/rio/cli/cmd/route"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/urfave/cli"
)

type Ps struct {
	C_Containers bool `desc:"print containers, not services"`
	A_All        bool `desc:"print all resources, including router and externalservice"`
}

func (p *Ps) Customize(cmd *cli.Command) {
	cmd.Flags = append(table.WriterFlags(), cmd.Flags...)
}

func (p *Ps) Run(ctx *clicontext.CLIContext) error {
	args := ctx.CLI.Args()

	if len(args) == 0 {
		if p.C_Containers {
			return Containers(ctx)
		}
		if p.A_All {
			return p.showAll(ctx)
		}
		return p.apps(ctx)
	}

	isRevision := true
	for _, arg := range args {
		if strings.Contains(arg, ":") {
			isRevision = false
			continue
		}
		if !isRevision {
			return errors.New("Can not pass both service and revision")
		}
	}

	if isRevision && !p.C_Containers {
		return revision.Revisions(ctx)
	}

	if p.C_Containers {
		return Containers(ctx)
	}

	return Pods(ctx)
}

func (p *Ps) showAll(ctx *clicontext.CLIContext) error {
	buffer := &strings.Builder{}
	ctx.WithWriter(buffer)

	fmt.Fprintf(buffer, "Applications\n")
	if err := p.apps(ctx); err != nil {
		return err
	}
	fmt.Fprintf(buffer, "\n")

	fmt.Fprintf(buffer, "Routers\n")
	if err := route.ListRouters(ctx); err != nil {
		return err
	}
	fmt.Fprintf(buffer, "\n")

	fmt.Fprintf(buffer, "ExternalServices\n")
	if err := externalservice.ListExternalServices(ctx); err != nil {
		return err
	}
	fmt.Fprintf(buffer, "\n")

	fmt.Print(buffer.String())
	return nil
}
