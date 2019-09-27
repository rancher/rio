package promote

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
)

type Promote struct {
	RolloutIncrement int  `desc:"Rollout increment value" default:"5"`
	RolloutInterval  int  `desc:"Rollout interval value" default:"5"`
	Rollout          bool `desc:"Whether to rollout gradually. Default to true" default:"true"`
}

func (p *Promote) Run(ctx *clicontext.CLIContext) error {
	return nil
}
