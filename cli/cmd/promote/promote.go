package promote

import (
	"errors"
	"fmt"
	"time"

	"github.com/rancher/mapper"
	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Promote struct {
	Increment int  `desc:"Increment value" default:"5"`
	Interval  int  `desc:"Interval value" default:"5"`
	Pause     bool `desc:"Whether to pause rollout or continue it. Default to false" default:"false"`
}

func (p *Promote) Run(ctx *clicontext.CLIContext) error {
	ctx.NoPrompt = true
	var allErrors []error
	arg := ctx.CLI.Args()
	if !arg.Present() {
		return errors.New("at least one argument is needed")
	}
	serviceName := arg.First()
	svcs, err := util.ListAppServicesFromServiceName(ctx, serviceName)
	if err != nil {
		return err
	}

	for _, s := range svcs {
		err := ctx.UpdateResource(types.Resource{
			Namespace: s.Namespace,
			Name:      s.Name,
			Type:      types.ServiceType,
		}, func(obj runtime.Object) error {
			s := obj.(*riov1.Service)
			if s.Spec.Weight == nil {
				s.Spec.Weight = new(int)
			}
			s.Spec.RolloutConfig = &riov1.RolloutConfig{
				Pause:     p.Pause,
				Increment: p.Increment,
				Interval: metav1.Duration{
					Duration: time.Duration(p.Interval) * time.Second,
				},
			}
			if s.Name == serviceName {
				*s.Spec.Weight = 100
				fmt.Printf("%s promoted\n", s.Name)
			} else {
				*s.Spec.Weight = 0
			}
			return nil
		})
		allErrors = append(allErrors, err)
	}
	return mapper.NewErrors(allErrors...)
}
