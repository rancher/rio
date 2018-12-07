package publicdomain

import (
	"strings"

	"github.com/rancher/rio/cli/pkg/lookup"

	"github.com/rancher/norman/types"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/types/client/project/v1"
)

type Rm struct {
}

func (r *Rm) Run(ctx *clicontext.CLIContext) error {
	cluster, err := ctx.Cluster()
	if err != nil {
		return err
	}
	spaceClient, err := cluster.Client()
	if err != nil {
		return err
	}
	publicDomains, err := spaceClient.PublicDomain.List(&types.ListOpts{})
	if err != nil {
		return err
	}
	m := make(map[string]*client.PublicDomain, 0)
	for _, p := range publicDomains.Data {
		m[p.DomainName] = &p
	}
	for _, arg := range ctx.CLI.Args() {
		if toDelete, ok := m[strings.TrimPrefix(arg, "https://")]; ok {
			if err := spaceClient.PublicDomain.Delete(toDelete); err != nil {
				return err
			}
		} else {
			resource, err := lookup.Lookup(ctx, arg, client.PublicDomainType)
			if err != nil {
				return err
			}

			client, err := ctx.ClientLookup(resource.Type)
			if err != nil {
				return err
			}

			err = client.Delete(&resource.Resource)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
