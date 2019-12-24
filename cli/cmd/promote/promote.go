package promote

import (
	"errors"
	"fmt"
	"time"

	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/cmd/weight"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/services"
)

type Promote struct {
	Duration string `desc:"How long the rollout should take. An approximation, actual time may fluctuate" default:"0s"`
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
	svcs, err := util.ListAppServicesFromAppName(ctx, resource.Namespace, resource.App)
	if err != nil {
		return err
	}
	svc := resource.Object.(*riov1.Service)
	promoteWeight, rolloutConfig, err := services.GenerateWeightAndRolloutConfig(svc, svcs, 100, duration, p.Pause)
	if err != nil {
		return err
	}
	err = weight.PromoteService(ctx, resource, rolloutConfig, promoteWeight)
	if err != nil {
		return err
	}
	id, _ := table.FormatID(resource.Object, resource.Namespace)
	fmt.Printf("%s promoted\n", id)
	return nil
}
