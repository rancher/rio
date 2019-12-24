package weight

import (
	"errors"
	"fmt"
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

type Weight struct {
	Duration string `desc:"How long the rollout should take. An approximation, actual time may fluctuate" default:"0s"`
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
	svcs, err := util.ListAppServicesFromAppName(ctx, resource.Namespace, resource.App)
	if err != nil {
		return err
	}
	svc := resource.Object.(*riov1.Service)
	weight, rolloutConfig, err := services.GenerateWeightAndRolloutConfig(svc, svcs, target, duration, w.Pause)
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

// PromoteService sets one svc to weight 100% and all others to 0%. If the svc doesn't need updating it is skipped to avoid conflicts.
func PromoteService(ctx *clicontext.CLIContext, resource types.Resource, rolloutConfig *riov1.RolloutConfig, promoteWeight int) error {
	var allErrors []error
	svcs, err := util.ListAppServicesFromAppName(ctx, resource.Namespace, resource.App)
	if err != nil {
		return err
	}
	if len(svcs) == 0 {
		return errors.New("no services found")
	}
	for _, s := range svcs {
		app, version := services.AppAndVersion(s)
		if (version == resource.Version && promotedSvcNeedsUpdate(s, promoteWeight, rolloutConfig) == true) || (version != resource.Version && s.Status.ComputedWeight != nil && *s.Status.ComputedWeight > 0) {
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

// Returns true if promoted service needs an update, only true if something needs to change
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
