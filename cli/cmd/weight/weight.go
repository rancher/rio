package weight

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/rio/cli/cmd/promote"
	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
	"github.com/rancher/wrangler/pkg/merr"
	"k8s.io/apimachinery/pkg/runtime"
)

type Weight struct {
	Increment int  `desc:"Amount of weight to increment on each interval" default:"5"`
	Interval  int  `desc:"Interval seconds between each increment" default:"5"`
	Pause     bool `desc:"Whether to pause rollout or continue it. Default to false" default:"false"`
}

func (w *Weight) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return errors.New("at least one parameter is required. Run -h to see options")
	}
	ctx.NoPrompt = true
	rolloutConfig := &riov1.RolloutConfig{
		Pause:           w.Pause,
		Increment:       w.Increment,
		IntervalSeconds: w.Interval,
	}
	for _, arg := range ctx.CLI.Args() {
		if strings.Contains(arg, "%") {
			if len(ctx.CLI.Args()) > 1 {
				return errors.New("only one percentage setting allowed per weight command")
			}
			return setPercentageWeight(ctx, ctx.CLI.Args()[0], rolloutConfig)
		}
	}
	return setSpecWeight(ctx, ctx.CLI.Args(), rolloutConfig)
}

func setSpecWeight(ctx *clicontext.CLIContext, args []string, rolloutConfig *riov1.RolloutConfig) error {
	var errs []error
	for _, arg := range args {
		serviceName, scale, err := validateArg(arg)
		if err != nil {
			return err
		}
		resource, err := ctx.ByID(serviceName)
		if err != nil {
			return err
		}
		err = ctx.UpdateResource(resource, func(obj runtime.Object) error {
			service := obj.(*riov1.Service)
			if service.Spec.Weight == nil {
				service.Spec.Weight = new(int)
			}
			*service.Spec.Weight = scale
			service.Spec.RolloutConfig = rolloutConfig
			return nil
		})
		errs = append(errs, err)
	}
	return merr.NewErrors(errs...)
}

func setPercentageWeight(ctx *clicontext.CLIContext, arg string, rolloutConfig *riov1.RolloutConfig) error {
	serviceName, scale, err := validateArg(arg)
	if err != nil {
		return err
	}
	resource, err := ctx.ByID(serviceName)
	if err != nil {
		return err
	}
	if scale == 100 {
		return promote.PerformPromote(ctx, serviceName, rolloutConfig)
	}
	svcs, err := util.ListAppServicesFromServiceName(ctx, serviceName)
	if err != nil {
		return err
	}
	// first find all weight on other versions of this service
	otherSvcTotalWeight := 0
	svc := resource.Object.(*riov1.Service)
	for _, s := range svcs {
		if s.Name == svc.Name {
			continue
		}
		if s.Status.ComputedWeight != nil && *s.Status.ComputedWeight > 0 {
			otherSvcTotalWeight += *s.Status.ComputedWeight
		}
	}
	// now calculate what computed weight should be to hit or percentage target, and set that
	weight := scale
	if otherSvcTotalWeight > 0 {
		weight = int(float64(otherSvcTotalWeight)/(1-(float64(scale)/100))) - otherSvcTotalWeight
	}
	return ctx.UpdateResource(resource, func(obj runtime.Object) error {
		s := obj.(*riov1.Service)
		if s.Spec.Weight == nil {
			s.Spec.Weight = new(int)
		}
		*s.Spec.Weight = weight
		s.Spec.RolloutConfig = rolloutConfig
		return nil
	})
}

func validateArg(arg string) (string, int, error) {
	serviceName, scaleStr := kv.Split(arg, "=")
	scaleStr = strings.TrimSuffix(scaleStr, "%")
	if scaleStr == "" {
		return serviceName, 0, errors.New("weight params must be in the format of SERVICE=WEIGHT, for example: myservice=10% or myservice=20")
	}
	scale, err := strconv.Atoi(scaleStr)
	if err != nil {
		return serviceName, scale, fmt.Errorf("failed to parse %s: %v", arg, err)
	}
	if scale > 100 {
		return serviceName, scale, fmt.Errorf("scale cannot exceed 100")
	}
	return serviceName, scale, nil
}
