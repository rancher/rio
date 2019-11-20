package weight

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/wrangler/pkg/kv"
	"github.com/rancher/wrangler/pkg/merr"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	DefaultInterval = 4
	PromoteWeight   = 10000
)

type Weight struct {
	Duration string `desc:"How long the rollout should take" default:"0s"`
	Pause    bool   `desc:"Whether to pause rollout or continue it. Default to false" default:"false"`
}

func (w *Weight) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return errors.New("at least one parameter is required. Run -h to see options")
	}
	ctx.NoPrompt = true
	if len(ctx.CLI.Args()) > 1 {
		for _, arg := range ctx.CLI.Args() {
			if strings.Contains(arg, "%") {
				return errors.New("only one percentage setting allowed per weight command")
			}
		}
	}
	duration, err := time.ParseDuration(w.Duration)
	if err != nil {
		return err
	}
	serviceName, target, err := validateArg(ctx.CLI.Args().First())
	if err != nil {
		return err
	}
	resource, err := ctx.ByID(serviceName)
	if err != nil {
		return err
	}
	weight, rolloutConfig, err := GenerateWeightAndRolloutConfig(ctx, resource, target, duration, w.Pause)
	if err != nil {
		return err
	}
	if target == 100 {
		return PromoteService(ctx, resource, rolloutConfig, weight)
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

func GenerateWeightAndRolloutConfig(ctx *clicontext.CLIContext, obj types.Resource, targetPercentage int, duration time.Duration, pause bool) (int, *riov1.RolloutConfig, error) {
	if duration.Hours() > 10 {
		return 0, nil, errors.New("cannot perform rollout longer than 10 hours") // over 10 hours we go under increment of 1/10k, given 2 second. Also see safety valve below in increment.
	}
	svcs, err := util.ListAppServicesFromAppName(ctx, obj.Namespace, obj.App)
	if err != nil && err.Error() == "no services found" || len(svcs) == 0 { // todo: make this string check less brittle
		return targetPercentage * 100, &riov1.RolloutConfig{}, nil
	} else if err != nil {
		return 0, nil, err
	}

	currComputedWeight := 0
	if svc, ok := obj.Object.(*riov1.Service); ok {
		if svc.Status.ComputedWeight != nil && *svc.Status.ComputedWeight > 0 {
			currComputedWeight = *svc.Status.ComputedWeight
		}
	}

	totalCurrWeight := 0
	for _, s := range svcs {
		if s.Status.ComputedWeight != nil && *s.Status.ComputedWeight > 0 {
			totalCurrWeight += *s.Status.ComputedWeight
		}
	}
	if targetPercentage == CalcWeightPercentage(currComputedWeight, totalCurrWeight) {
		return 0, nil, errors.New("cannot rollout, already at target percentage")
	}
	totalCurrWeightOtherSvcs := totalCurrWeight - currComputedWeight
	newComputedWeight := calcComputedWeight(targetPercentage, totalCurrWeightOtherSvcs)

	// if not immediate rollout figure out increment
	increment := 0
	if duration.Seconds() >= 2.0 {
		increment, err = calcIncrement(duration, targetPercentage, totalCurrWeight, totalCurrWeightOtherSvcs)
		if err != nil {
			return 0, nil, err
		}
	}
	rolloutConfig := &riov1.RolloutConfig{
		Pause:           pause,
		Increment:       increment,
		IntervalSeconds: DefaultInterval,
	}
	return newComputedWeight, rolloutConfig, nil
}

// PromoteService sets one svc to weight 100% and all others to 0%. If the svc doesn't need updating it is skipped to avoid conflicts.
func PromoteService(ctx *clicontext.CLIContext, resource types.Resource, rolloutConfig *riov1.RolloutConfig, promoteWeight int) error {
	var allErrors []error
	svcs, err := util.ListAppServicesFromAppName(ctx, resource.Namespace, resource.App)
	if err != nil {
		return err
	}
	for _, s := range svcs {
		app, version := services.AppAndVersion(&s)
		if (version == resource.Version && promotedSvcNeedsUpdate(&s, promoteWeight, rolloutConfig) == true) || (version != resource.Version && s.Status.ComputedWeight != nil && *s.Status.ComputedWeight > 0) {
			err := ctx.UpdateResource(types.Resource{
				Namespace: s.Namespace,
				Name:      s.Name,
				App:       app,
				Version:   version,
				Type:      types.ServiceType,
			}, func(obj runtime.Object) error {
				s := obj.(*riov1.Service)
				s.Spec.RolloutConfig = rolloutConfig
				if s.Spec.Weight == nil {
					s.Spec.Weight = new(int)
				}
				if version == resource.Version {
					*s.Spec.Weight = promoteWeight
				} else {
					*s.Spec.Weight = 0
				}
				return nil
			})
			allErrors = append(allErrors, err)
		}
	}
	return merr.NewErrors(allErrors...)
}

func validateArg(arg string) (string, int, error) {
	serviceName, scaleStr := kv.Split(arg, "=")
	scaleStr = strings.TrimSuffix(scaleStr, "%")
	if scaleStr == "" {
		return serviceName, 0, errors.New("weight params must be in the format of SERVICE=WEIGHT, for example: myservice=10%")
	}
	scale, err := strconv.Atoi(scaleStr)
	if err != nil {
		return serviceName, scale, fmt.Errorf("failed to parse %s: %v", arg, err)
	}
	if scale > 100 {
		return serviceName, scale, fmt.Errorf("weight cannot exceed 100")
	}
	return serviceName, scale, nil
}

// Only update promoted service if something needs to change
// `rio run` cmd sets details and then promote call comes in immediately after
func promotedSvcNeedsUpdate(svc *riov1.Service, weight int, rc *riov1.RolloutConfig) bool {
	if svc.Spec.Weight == nil {
		return true
	}
	if *svc.Spec.Weight != weight {
		return true
	}
	if !reflect.DeepEqual(rc, svc.Spec.RolloutConfig) {
		return true
	}
	return false

}

// Get curr weight as percentage, rounded to nearest percent
func CalcWeightPercentage(weight, totalWeight int) int {
	if totalWeight == 0 || weight == 0 {
		return 0
	}
	return int(math.Round(float64(weight) / float64(totalWeight) / 0.01))
}

// Find the weight that would hit our target percentage without touching other service weights
// ie: if 2 svcs at 50/50 and you want one at 75%, newComputedWeight would be 150
func calcComputedWeight(targetPercentage int, totalCurrWeightOtherSvcs int) int {
	if targetPercentage == 100 {
		return PromoteWeight
	} else if totalCurrWeightOtherSvcs > 0 {
		return int(float64(totalCurrWeightOtherSvcs)/(1-(float64(targetPercentage)/100))) - totalCurrWeightOtherSvcs
	}
	return targetPercentage
}

// Determine increment we should step by given duration
// Note that we don't care (because blind to direction of scaling) if increment is larger than newComputedWeight, rollout controller will handle overflow case
func calcIncrement(duration time.Duration, targetPercentage, totalCurrWeight, totalCurrWeightOtherSvcs int) (int, error) {
	steps := duration.Seconds() / float64(DefaultInterval) // First get rough amount of steps we want to take
	if steps < 1.0 {
		steps = 1.0
	}
	newComputedWeight := calcComputedWeight(targetPercentage, totalCurrWeightOtherSvcs)
	totalNewWeight := totalCurrWeightOtherSvcs + newComputedWeight // Given the future total weight which includes our newWeight...
	difference := totalNewWeight - totalCurrWeight                 // Find the difference between future total weight and current total weight
	if targetPercentage == 100 {
		difference = PromoteWeight // In this case the future total weight is now only our newComputedWeight and difference is always 1k
	}
	if difference == 0 {
		return 0, nil // if there is no difference return now so we don't error out below
	}
	increment := int(math.Abs(math.Round(float64(difference) / steps))) // Divide by steps to get rough increment
	if increment == 0 {                                                 // Error out if increment was below 1, and thus rounded to 0
		return 0, errors.New("unable to perform rollout, given duration too long for current weight")
	}
	return increment, nil

}
