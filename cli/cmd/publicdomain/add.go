package publicdomain

import (
	"fmt"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/types/client/space/v1beta1"
)

type Add struct {
	Space    string `desc:"The space name of target routeset or service"`
	Stack    string `desc:"The stack name of target routeset or service"`
	Domain   string `desc:"The domain name to override"`
	Service  string `desc:"The service target"`
	RouteSet string `desc:"The routeset target"`
}

func (a *Add) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) != 1 {
		return fmt.Errorf("Incorrect Usage. Example: rio domain add $Name")
	}
	cluster, err := ctx.Cluster()
	if err != nil {
		return err
	}
	spaceClient, err := cluster.Client()
	if err != nil {
		return err
	}
	domain := &client.PublicDomain{
		Name:         ctx.CLI.Args().Get(0),
		SpaceName:    a.Space,
		StackName:    a.Stack,
		ServiceName:  a.Service,
		DomainName:   a.Domain,
		RouteSetName: a.RouteSet,
	}
	domain, err = spaceClient.PublicDomain.Create(domain)
	if err != nil {
		return err
	}
	fmt.Println(domain.Name)
	return nil
}
