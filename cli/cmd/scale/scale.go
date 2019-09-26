package scale

import (
	"fmt"
	"strconv"
	"strings"

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
		err := ctx.Update(name, clitypes.ServiceType, func(obj runtime.Object) error {
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
				if service.Spec.Autoscale.Concurrency == nil {
					service.Spec.Autoscale.Concurrency = &[]int{10}[0]
				}
				service.Spec.Autoscale.MinReplicas = &minScale
				service.Spec.Autoscale.MaxReplicas = &maxScale
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
				service.Spec.Autoscale.MinReplicas = nil
				service.Spec.Autoscale.MaxReplicas = nil
				service.Status.ObservedScale = nil
			}

			return nil
		})
		errors = append(errors, err)
	}

	return mapper.NewErrors(errors...)
}
