package weight

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/rio/cli/pkg/stack"

	"github.com/rancher/mapper"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/types"
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
	if len(ctx.CLI.Args()) < 2 {
		return errors.New("at least two parameters are required. Run -h to see options")
	}

	namespace, name := stack.NamespaceAndName(ctx, ctx.CLI.Args()[0])
	app, err := ctx.Rio.Apps(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	versionMap := map[string]string{}
	var totalWeight int
	for _, rev := range app.Spec.Revisions {
		totalWeight += rev.Weight
		versionMap[rev.Version] = fmt.Sprintf("%s/%s", app.Namespace, rev.ServiceName)
	}

	var errors []error
	for _, arg := range ctx.CLI.Args()[1:] {
		version, scaleStr := kv.Split(arg, "=")
		var weightToSet int
		if !strings.HasSuffix(scaleStr, "%") {
			scale, err := strconv.Atoi(scaleStr)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %v", arg, err)
			}
			weightToSet = scale
		} else {
			scaleStr = strings.TrimSuffix(scaleStr, "%")

			if scaleStr == "" {
				return fmt.Errorf("weight params must be in the format of SERVICE=PERCENTAGE, for example: myservice=10%")
			}
			scale, err := strconv.Atoi(scaleStr)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %v", arg, err)
			}
			weightToSet = int(scale * totalWeight / 100.0)
		}

		resource, err := lookup.Lookup(ctx, versionMap[version], types.ServiceType)
		if err != nil {
			return err
		}

		err = ctx.UpdateResource(resource, func(obj runtime.Object) error {
			service := obj.(*v1.Service)
			service.Spec.ServiceRevision.Weight = weightToSet
			if w.NoRollout {
				service.Spec.Rollout = false
			} else {
				service.Spec.RolloutInterval = w.RolloutInterval
				service.Spec.RolloutIncrement = w.RolloutIncrement
			}
			return nil
		})
		errors = append(errors, err)
	}

	return mapper.NewErrors(errors...)
}
