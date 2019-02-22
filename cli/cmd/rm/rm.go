package rm

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Rm struct {
	T_Type string `desc:"delete specific type"`
}

func (r *Rm) Run(ctx *clicontext.CLIContext) error {
	types := []string{clitypes.ServiceType, clitypes.StackType, clitypes.PodType, clitypes.ConfigType, clitypes.RouteSetType, clitypes.VolumeType, clitypes.ExternalServiceType}
	if len(r.T_Type) > 0 {
		types = []string{r.T_Type}
	}

	return Remove(ctx, types...)
}

func Remove(ctx *clicontext.CLIContext, types ...string) error {
	// todo: add waiter
	client, err := ctx.KubeClient()
	if err != nil {
		return err
	}

	for _, arg := range ctx.CLI.Args() {
		resource, err := lookup.Lookup(ctx, arg, types...)
		if err != nil {
			return err
		}

		switch resource.Type {
		case clitypes.ServiceType:
			if err := client.Rio.Services(resource.Namespace).Delete(resource.Name, &metav1.DeleteOptions{}); err != nil {
				return err
			}
		case clitypes.StackType:
			if err := client.Rio.Stacks(resource.Namespace).Delete(resource.Name, &metav1.DeleteOptions{}); err != nil {
				return err
			}
		case clitypes.PodType:
			if err := client.Core.Pods(resource.Namespace).Delete(resource.Name, &metav1.DeleteOptions{}); err != nil {
				return err
			}
		case clitypes.ConfigType:
			if err := client.Rio.Configs(resource.Namespace).Delete(resource.Name, &metav1.DeleteOptions{}); err != nil {
				return err
			}
		case clitypes.RouteSetType:
			if err := client.Rio.Services(resource.Namespace).Delete(resource.Name, &metav1.DeleteOptions{}); err != nil {
				return err
			}
		case clitypes.VolumeType:
			if err := client.Rio.Volumes(resource.Namespace).Delete(resource.Name, &metav1.DeleteOptions{}); err != nil {
				return err
			}
		case clitypes.ExternalServiceType:
			if err := client.Rio.ExternalServices(resource.Namespace).Delete(resource.Name, &metav1.DeleteOptions{}); err != nil {
				return err
			}
		}

	}

	return nil
}
