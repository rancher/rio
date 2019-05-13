package weight

import (
	"errors"
	err2 "errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/mapper"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/stack"
	"github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Weight struct {
	RolloutIncrement int  `desc:"Rollout increment value" default:"5"`
	RolloutInterval  int  `desc:"Rollout interval value" default:"5"`
	NoRollout        bool `desc:"Don't rollout"`
}

func (w *Weight) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return errors.New("at least one parameter is required. Run -h to see options")
	}

	return ScaleAndAllocate(ctx, ctx.CLI.Args(), w.NoRollout, w.RolloutIncrement, w.RolloutInterval)
}

func ScaleAndAllocate(ctx *clicontext.CLIContext, args []string, noRollout bool, increment, interval int) error {
	var errors []error
	appVersion, _ := kv.Split(ctx.CLI.Args()[0], "=")
	app, _ := kv.Split(appVersion, ":")

	namespace, name := stack.NamespaceAndName(ctx, app)
	appObj, err := ctx.Rio.Apps(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	versionMap := map[string]string{}
	for _, rev := range appObj.Spec.Revisions {
		versionMap[rev.Version] = fmt.Sprintf("%s/%s", appObj.Namespace, rev.ServiceName)
	}

	toSet := map[string]struct{}{}
	reminder := 100
	for _, arg := range args {
		appVersion, scaleStr := kv.Split(arg, "=")
		_, version := kv.Split(appVersion, ":")
		toSet[version] = struct{}{}

		scaleStr = strings.TrimSuffix(scaleStr, "%")
		if scaleStr == "" {
			return err2.New("weight params must be in the format of SERVICE=PERCENTAGE, for example: myservice=10%")
		}
		scale, err := strconv.Atoi(scaleStr)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %v", arg, err)
		}
		if scale > 100 || reminder < 0 {
			return fmt.Errorf("scale can't not exceed 100")
		}

		// set other version weight to zero and break loop
		resource, err := lookup.Lookup(ctx, versionMap[version], types.ServiceType)
		if err != nil {
			return err
		}
		err = ctx.UpdateResource(resource, func(obj runtime.Object) error {
			service := obj.(*v1.Service)
			service.Spec.ServiceRevision.Weight = scale
			if noRollout {
				service.Spec.Rollout = false
			} else {
				service.Spec.Rollout = true
				service.Spec.RolloutInterval = interval
				service.Spec.RolloutIncrement = increment
			}
			return nil
		})
		errors = append(errors, err)
		reminder -= scale
	}

	var toAllocate []riov1.Revision
	total := 0
	for _, rev := range appObj.Spec.Revisions {
		if _, ok := toSet[rev.Version]; ok {
			continue
		}
		total += rev.AdjustedWeight
		toAllocate = append(toAllocate, rev)
	}

	added := 0
	for i, rev := range toAllocate {
		resource, err := lookup.Lookup(ctx, versionMap[rev.Version], types.ServiceType)
		if err != nil {
			return err
		}
		weight := 0
		if i == len(toAllocate)-1 {
			weight = reminder - added
		} else {
			weight = int(float64(rev.AdjustedWeight) / float64(total) * float64(reminder))
			added += weight
		}
		err = ctx.UpdateResource(resource, func(obj runtime.Object) error {
			service := obj.(*v1.Service)
			service.Spec.ServiceRevision.Weight = weight
			if noRollout {
				service.Spec.Rollout = false
			} else {
				service.Spec.Rollout = true
				service.Spec.RolloutInterval = interval
				service.Spec.RolloutIncrement = increment
			}
			return nil
		})
		errors = append(errors, err)
	}

	return mapper.NewErrors(errors...)
}
