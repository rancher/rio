package promote

import (
	"errors"
	"time"

	"github.com/rancher/rio/cli/cmd/weight"
	"github.com/rancher/rio/cli/pkg/clicontext"
)

type Promote struct {
	Duration string `desc:"How long the rollout should take" default:"0s"`
	Pause    bool   `desc:"Whether to pause rollout or continue it. Default to false" default:"false"`
}

func (p *Promote) Run(ctx *clicontext.CLIContext) error {
	ctx.NoPrompt = true
	arg := ctx.CLI.Args()
	if !arg.Present() {
		return errors.New("at least one argument is needed")
	}
	serviceName := arg.First()
	resource, err := ctx.ByID(serviceName)
	if err != nil {
		return err
	}
	duration, err := time.ParseDuration(p.Duration)
	if err != nil {
		return err
	}
	promoteWeight, rolloutConfig, err := weight.GenerateWeightAndRolloutConfig(ctx, resource, 100, duration, p.Pause)
	if err != nil {
		return err
	}
	return weight.PromoteService(ctx, resource, rolloutConfig, promoteWeight)
}
