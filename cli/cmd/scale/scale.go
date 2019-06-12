package scale

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/up/questions"
	"github.com/rancher/mapper"
	"github.com/rancher/rio/cli/pkg/clicontext"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
	"k8s.io/apimachinery/pkg/runtime"
)

type Scale struct {
}

func (s *Scale) Run(ctx *clicontext.CLIContext) error {
	var errors []error

	for _, arg := range ctx.CLI.Args() {
		name, scaleStr := kv.Split(arg, "=")
		if !strings.Contains(name, ":") {
			app, err := lookup.Lookup(ctx, name, clitypes.AppType)
			if err != nil {
				return err
			}
			var versions []string
			for _, rev := range app.Object.(*v1.App).Spec.Revisions {
				if rev.AdjustedWeight == 0 {
					continue
				}
				versions = append(versions, rev.Version)
			}
			if len(versions) == 1 {
				name = name + ":" + versions[0]
			} else {
				var options []string
				for i, ver := range versions {
					options = append(options, fmt.Sprintf("[%v] %v\n", i+1, ver))
				}
				num, err := questions.PromptOptions("Choose which version to scale\n", -1, options...)
				if err != nil {
					return err
				}
				name = name + ":" + versions[num]
			}
		}
		err := ctx.Update(name, clitypes.ServiceType, func(obj runtime.Object) error {
			service := obj.(*v1.Service)

			if strings.ContainsRune(scaleStr, '-') {
				min, max := kv.Split(scaleStr, "-")
				minScale, _ := strconv.Atoi(min)
				maxScale, _ := strconv.Atoi(max)
				if service.Spec.AutoscaleConfig.Concurrency == nil {
					service.Spec.AutoscaleConfig.Concurrency = &[]int{10}[0]
				}
				service.Spec.AutoscaleConfig.MinScale = &minScale
				service.Spec.AutoscaleConfig.MaxScale = &maxScale
			} else {
				scale, err := strconv.Atoi(scaleStr)
				if err != nil {
					return fmt.Errorf("failed to parse %s: %v", arg, err)
				}
				service.Spec.Scale = scale
				service.Spec.MinScale = nil
				service.Spec.MaxScale = nil
				service.Status.ObservedScale = nil
			}

			return nil
		})
		errors = append(errors, err)
	}

	return mapper.NewErrors(errors...)
}
