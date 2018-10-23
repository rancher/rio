package publicdomain

import (
	"fmt"

	"github.com/rancher/norman/types"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/types/client/space/v1beta1"
)

type Ls struct {
}

func (l *Ls) Run(ctx *clicontext.CLIContext) error {
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
	writer := table.NewWriter([][]string{
		{"Name", "{{.Name}}"},
		{"Domain", "{{.DomainName}}"},
		{"Target", "{{. | formatTarget}}"},
	}, ctx)
	defer writer.Close()

	writer.AddFormatFunc("formatTarget", FormatTarget)
	for _, publicDomain := range publicDomains.Data {
		writer.Write(&publicDomain)
	}
	return writer.Err()
}

func FormatTarget(obj interface{}) (string, error) {
	if v, ok := obj.(*client.PublicDomain); ok {
		target := fmt.Sprintf("%s, %s, %s/", v.SpaceName, v.SpaceName, v.StackName)
		if v.ServiceName != "" {
			target += v.ServiceName
		} else if v.RouteSetName != "" {
			target += v.RouteSetName
		}
		return target, nil
	}
	return "", nil
}
