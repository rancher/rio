package promote

import (
	"fmt"

	"github.com/rancher/mapper"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/types"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Promote struct {
	Scale int `desc:"scale of service after promotion"`
}

func (p *Promote) Run(ctx *clicontext.CLIContext) error {
	var errors []error
	for _, arg := range ctx.CLI.Args() {
		err := ctx.Update(arg, types.ServiceType, func(obj runtime.Object) error {
			service := obj.(*v1.Service)
			if service.Spec.Revision.ParentService == "" {
				return fmt.Errorf("can not promote the base version")
			}

			service.Spec.Revision.Promote = true

			if p.Scale > 0 {
				service.Spec.Scale = p.Scale
			}

			return nil
		})
		errors = append(errors, err)
	}

	return mapper.NewErrors(errors...)
}
