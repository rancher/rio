package promote

import (
	"errors"

	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/cmd/weight"
	"github.com/rancher/rio/cli/pkg/clicontext"
)

type Promote struct {
	Increment int  `desc:"Percentage of weight to increment on each interval" default:"5"`
	Interval  int  `desc:"Interval seconds between each increment" default:"0"`
	Pause     bool `desc:"Whether to pause rollout or continue it. Default to false" default:"false"`
}

func (p *Promote) Run(ctx *clicontext.CLIContext) error {
	ctx.NoPrompt = true
	arg := ctx.CLI.Args()
	if !arg.Present() {
		return errors.New("at least one argument is needed")
	}
	serviceName := arg.First()
	svcs, err := util.ListAppServicesFromServiceName(ctx, serviceName)
	if err != nil {
		return err
	}
	rolloutConfig := weight.GenerateAppRolloutConfig(svcs, p.Pause, p.Increment, p.Interval)
	return weight.PromoteService(ctx, serviceName, rolloutConfig)
}
