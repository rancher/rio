package promote

import (
	"errors"
	"fmt"

	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/merr"
	"k8s.io/apimachinery/pkg/runtime"
)

type Promote struct {
	Increment int  `desc:"Amount of weight to increment on each interval" default:"5"`
	Interval  int  `desc:"Interval seconds between each increment" default:"5"`
	Pause     bool `desc:"Whether to pause rollout or continue it. Default to false" default:"false"`
}

func (p *Promote) Run(ctx *clicontext.CLIContext) error {
	ctx.NoPrompt = true
	arg := ctx.CLI.Args()
	if !arg.Present() {
		return errors.New("at least one argument is needed")
	}
	serviceName := arg.First()
	rolloutConfig := &riov1.RolloutConfig{
		Pause:           p.Pause,
		Increment:       p.Increment,
		IntervalSeconds: p.Interval,
	}
	return PerformPromote(ctx, serviceName, rolloutConfig)
}

func PerformPromote(ctx *clicontext.CLIContext, serviceName string, rolloutConfig *riov1.RolloutConfig) error {
	var allErrors []error
	svcs, err := util.ListAppServicesFromServiceName(ctx, serviceName)
	if err != nil {
		return err
	}
	versionToPromote := ctx.ParseID(serviceName).Version
	if versionToPromote == "" {
		return errors.New("invalid version specified")
	}
	for _, s := range svcs {
		err := ctx.UpdateResource(types.Resource{
			Namespace: s.Namespace,
			Name:      s.Name,
			App:       s.Spec.App,
			Version:   s.Spec.Version,
			Type:      types.ServiceType,
		}, func(obj runtime.Object) error {
			s := obj.(*riov1.Service)
			if s.Spec.Weight == nil {
				s.Spec.Weight = new(int)
			}
			s.Spec.RolloutConfig = rolloutConfig
			if s.Spec.Version == versionToPromote {
				*s.Spec.Weight = 100
				fmt.Printf("%s promoted\n", s.Name)
			} else {
				*s.Spec.Weight = 0
			}
			return nil
		})
		allErrors = append(allErrors, err)
	}
	return merr.NewErrors(allErrors...)
}
