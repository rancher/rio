package stacks

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/tables"
	"github.com/rancher/rio/cli/pkg/types"
)

type ls struct {
}

func (l *ls) Run(ctx *clicontext.CLIContext) error {
	stacks, err := ctx.List(types.StackType)
	if err != nil {
		return err
	}

	writer := tables.NewStack(ctx)
	return writer.Write(stacks)
}
