package ps

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/rio/cli/cmd/revision"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/urfave/cli"
)

type Ps struct {
	C_Containers bool `desc:"print containers, not services"`
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
