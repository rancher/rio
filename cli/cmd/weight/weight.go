package weight

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/mapper"
	"github.com/rancher/rio/cli/pkg/clicontext"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
	"k8s.io/apimachinery/pkg/runtime"
)

type Weight struct {
	Increment int  `desc:"Amount of weight to increment each interval" default:"5"`
	Interval  int  `desc:"Interval seconds between each increment" default:"5"`
	Pause     bool `desc:"Whether to pause rollout or continue it. Default to false" default:"false"`
}

func (w *Weight) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return errors.New("at least one parameter is required. Run -h to see options")
	}
	ctx.NoPrompt = true
	return scaleAndAllocate(ctx, ctx.CLI.Args(), w.Pause, w.Increment, w.Interval)
}

func scaleAndAllocate(ctx *clicontext.CLIContext, args []string, pause bool, increment int, interval int) error {
	var errs []error
	// todo: fixing below in https://github.com/rancher/rio/issues/682
	//serviceName, _ := kv.Split(ctx.CLI.Args()[0], "=")
	//svcs, err := util.ListAppServicesFromServiceName(ctx, serviceName)
	//if err != nil {
	//	return err
	//}

	//cmdSet := map[string]int{}
	reminder := 100

	// grab all services that already had weight allocated
	//var toAllocate []riov1.Service
	//total := 0
	//for _, s := range svcs {
	//	// Don't count ones specified in the weight cmd
	//	if _, ok := cmdSet[s.Name]; ok {
	//		continue
	//	}
	//	if s.Status.ComputedWeight != nil && *s.Status.ComputedWeight > 0 {
	//		total += *s.Status.ComputedWeight
	//		toAllocate = append(toAllocate, s)
	//	}
	//}

	// First update spec weight on anything specified in command
	for _, arg := range args {
		serviceName, scaleStr := kv.Split(arg, "=")
		//cmdSet[serviceName] = 1
		scaleStr = strings.TrimSuffix(scaleStr, "%")
		if scaleStr == "" {
			return errors.New("weight params must be in the format of SERVICE=PERCENTAGE, for example: myservice=10%")
		}
		scale, err := strconv.Atoi(scaleStr)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %v", arg, err)
		}
		if scale > 100 || reminder < 0 {
			return fmt.Errorf("scale can't not exceed 100")
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
			service.Spec.RolloutConfig = &riov1.RolloutConfig{
				Pause:           pause,
				Increment:       increment,
				IntervalSeconds: interval,
			}
			return nil
		})
		errs = append(errs, err)
		//reminder -= scale
	}

	// now allocate any remaining weight across those pre-weighted services
	//added := 0
	//for i, rev := range toAllocate {
	//	resource, err := ctx.ByID(rev.Name)
	//	if err != nil {
	//		return err
	//	}
	//	weight := 0
	//	if i == len(toAllocate)-1 {
	//		weight = reminder - added
	//	} else {
	//		weight = int(float64(*rev.Status.ComputedWeight) / float64(total) * float64(reminder))
	//		added += weight
	//	}
	//	err = ctx.UpdateResource(resource, func(obj runtime.Object) error {
	//		s := obj.(*riov1.Service)
	//		if s.Spec.Weight == nil {
	//			s.Spec.Weight = new(int)
	//		}
	//		*s.Spec.Weight = weight
	//		s.Spec.RolloutConfig = &riov1.RolloutConfig{
	//			Pause:           pause,
	//			Increment:       increment,
	//			IntervalSeconds: interval,
	//		}
	//		return nil
	//	})
	//
	//	errs = append(errs, err)
	//}

	return mapper.NewErrors(errs...)
}
