package weight

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/mapper"
	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
	"k8s.io/apimachinery/pkg/runtime"
)

type Weight struct {
	Increment int  `desc:"Percentage of weight to increment on each interval" default:"5"`
	Interval  int  `desc:"Interval seconds between each increment" default:"0"`
	Pause     bool `desc:"Whether to pause rollout or continue it. Default to false" default:"false"`
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
	serviceName, scale, err := validateArg(ctx.CLI.Args().First())
	if err != nil {
		return err
	}
	resource, err := ctx.ByID(serviceName)
	if err != nil {
		return err
	}
	svcs, err := util.ListAppServicesFromServiceName(ctx, serviceName)
	if err != nil {
		return err
	}
	rolloutConfig := GenerateAppRolloutConfig(svcs, w.Pause, w.Increment, w.Interval)
	if scale == 100 {
		return PromoteService(ctx, serviceName, rolloutConfig)
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

func GenerateAppRolloutConfig(svcs []riov1.Service, pause bool, increment int, interval int) *riov1.RolloutConfig {
	total := 0
	for _, s := range svcs {
		if s.Status.ComputedWeight != nil {
			total += *s.Status.ComputedWeight
		}
	}
	rolloutConfig := &riov1.RolloutConfig{
		Pause:           pause,
		Increment:       int(float64(increment) * (float64(total) / 100.0)),
		IntervalSeconds: interval,
	}
	return rolloutConfig
}

func PromoteService(ctx *clicontext.CLIContext, serviceName string, rolloutConfig *riov1.RolloutConfig) error {
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
		if s.Spec.Version == versionToPromote || (s.Status.ComputedWeight != nil && *s.Status.ComputedWeight > 0) { // don't update non-promoted svcs without weight allocated
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
					id, _ := table.FormatID(obj, s.Namespace)
					fmt.Printf("%s promoted\n", id)
				} else {
					*s.Spec.Weight = 0
				}
				return nil
			})
			allErrors = append(allErrors, err)
		}
	}
	return mapper.NewErrors(allErrors...)
}
