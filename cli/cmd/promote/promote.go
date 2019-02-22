package promote

import (
	"fmt"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Promote struct {
	Scale int `desc:"scale of service after promotion"`
}

func (p *Promote) Run(ctx *clicontext.CLIContext) error {
	client, err := ctx.KubeClient()
	if err != nil {
		return err
	}

	for _, arg := range ctx.CLI.Args() {
		resource, err := lookup.Lookup(ctx, arg, types.ServiceType)
		if err != nil {
			return err
		}

		service, err := client.Rio.Services(resource.Namespace).Get(resource.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if service.Spec.Revision.ParentService == "" {
			return fmt.Errorf("can not promote the base version")
		}

		service.Spec.Revision.Promote = true

		if p.Scale > 0 {
			service.Spec.Scale = p.Scale
		}

		service, err = client.Rio.Services(resource.Namespace).Update(service)
		if err != nil {
			return err
		}
	}

	return nil
}
