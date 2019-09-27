package weight

import (
	"errors"

	"github.com/rancher/rio/cli/pkg/clicontext"
)

type Weight struct {
	RolloutIncrement int  `desc:"Rollout increment value" default:"5"`
	RolloutInterval  int  `desc:"Rollout interval value" default:"5"`
	Rollout          bool `desc:"Whether to rollout gradually"`
}

func (w *Weight) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return errors.New("at least one parameter is required. Run -h to see options")
	}
	ctx.NoPrompt = true
	return nil
}
