package ps

import (
	"fmt"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	"github.com/urfave/cli"
)

type Ps struct {
	C_Containers bool `desc:"print containers, not services"`
	A_All        bool `desc:"include all container, not just ones related to services"`
	E_External   bool `desc:"print external services"`
}

func (p *Ps) Customize(cmd *cli.Command) {
	cmd.Flags = append(table.WriterFlags(), cmd.Flags...)
}

func (p *Ps) Run(ctx *clicontext.CLIContext) error {
	var (
		stacks    = map[string]bool{}
		notStacks []string
	)

	if !p.C_Containers {
		for _, arg := range ctx.CLI.Args() {
			stack, err := lookup.Lookup(ctx, arg, client.StackType)
			if err == nil {
				stacks[stack.ID] = true
			} else {
				notStacks = append(notStacks, arg)
			}
		}
	}

	if p.C_Containers || (len(ctx.CLI.Args()) > 0 && len(stacks) == 0) {
		return p.containers(ctx)
	}

	if len(notStacks) > 0 {
		return fmt.Errorf("failed to find stacks for %v", notStacks)
	}

	return p.services(ctx, stacks)
}
