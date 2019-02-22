package publicdomain

import (
	"strings"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/types"
	projectv1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Rm struct {
}

func (r *Rm) Run(ctx *clicontext.CLIContext) error {
	client, err := ctx.KubeClient()
	if err != nil {
		return err
	}
	publicDomains, err := client.Project.PublicDomains("").List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	m := make(map[string]*projectv1.PublicDomain, 0)
	for _, p := range publicDomains.Items {
		m[p.Spec.DomainName] = &p
	}
	for _, arg := range ctx.CLI.Args() {
		if toDelete, ok := m[strings.TrimPrefix(arg, "https://")]; ok {
			if err := client.Project.PublicDomains("").Delete(toDelete.Name, &metav1.DeleteOptions{}); err != nil {
				return err
			}
		} else {
			resource, err := lookup.Lookup(ctx, arg, types.PublicDomainType)
			if err != nil {
				return err
			}

			if err := client.Project.PublicDomains("").Delete(resource.Name, &metav1.DeleteOptions{}); err != nil {
				return err
			}
		}
	}
	return nil
}
