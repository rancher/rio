package scale

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/rio/cli/cmd/util"

	"github.com/rancher/mapper"
	"github.com/rancher/rio/cli/pkg/clicontext"
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
		err := ctx.Update(name, func(obj runtime.Object) error {
			service := obj.(*v1.Service)

			if strings.ContainsRune(scaleStr, '-') {
				min, max := kv.Split(scaleStr, "-")
				minScale, err := strconv.Atoi(min)
				if err != nil {
					return err
				}
				maxScale, err := strconv.Atoi(max)
				if err != nil {
					return err
				}
				if service.Spec.Autoscale == nil {
					service.Spec.Autoscale = &v1.AutoscaleConfig{}
				}
				if service.Spec.Autoscale.Concurrency == 0 {
					service.Spec.Autoscale.Concurrency = 10
				}
				service.Spec.Autoscale.MinReplicas = util.ToInt32(minScale)
				service.Spec.Autoscale.MaxReplicas = util.ToInt32(maxScale)
				if minScale != 0 {
					service.Spec.Replicas = &minScale
				} else {
					service.Spec.Replicas = &[]int{1}[0]
				}
			} else {
				scale, err := strconv.Atoi(scaleStr)
				if err != nil {
					return fmt.Errorf("failed to parse %s: %v", arg, err)
				}
				service.Spec.Replicas = &scale
				service.Spec.Autoscale = nil
			}

			return nil
		})
		errors = append(errors, err)
	}

	return mapper.NewErrors(errors...)
}
