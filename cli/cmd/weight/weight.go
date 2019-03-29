package weight

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/mapper"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/types"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
	"k8s.io/apimachinery/pkg/runtime"
)

type Weight struct {
}

func (w *Weight) Run(ctx *clicontext.CLIContext) error {
	var errors []error

	for _, arg := range ctx.CLI.Args() {
		name, scaleStr := kv.Split(arg, "=")
		scaleStr = strings.TrimSuffix(scaleStr, "%")

		if scaleStr == "" {
			return fmt.Errorf("weight params must be in the format of SERVICE=PERCENTAGE, for example: mystack/myservice=10%%")
		}
		scale, err := strconv.Atoi(scaleStr)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %v", arg, err)
		}

		resource, err := lookup.Lookup(ctx, name, types.ServiceType)
		if err != nil {
			return err
		}

		err = ctx.UpdateResource(resource, func(obj runtime.Object) error {
			service := obj.(*v1.Service)
			service.Spec.Revision.Weight = scale
			return nil
		})
		errors = append(errors, err)
	}

	return mapper.NewErrors(errors...)
}
