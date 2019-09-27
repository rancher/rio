package inspect

import (
	"errors"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/urfave/cli"
)

type Inspect struct {
}

func (i *Inspect) Customize(cmd *cli.Command) {
	for _, f := range table.WriterFlags() {
		if f.GetName() == "format" {
			sf := f.(cli.StringFlag)
			sf.Value = "yaml"
			cmd.Flags = append(cmd.Flags, sf)
		}
	}
}

func (i *Inspect) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return errors.New("at least one argument is required")
	}

	ctx.ByID()

	return nil
}
