package publicdomain

import (
	"fmt"
	"strings"

	"github.com/rancher/rio/pkg/namespace"

	"github.com/rancher/norman/types"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/clientcfg"
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

	w := wrapper{
		clusterClient: cluster,
		spaceClient:   spaceClient.Space,
	}
	writer.AddFormatFunc("formatTarget", w.FormatTarget)
	for _, publicDomain := range publicDomains.Data {
		if publicDomain.RequestTLSCert {
			publicDomain.DomainName = "https://" + publicDomain.DomainName
		} else {
			publicDomain.DomainName = "http://" + publicDomain.DomainName
		}
		writer.Write(&publicDomain)
	}
	return writer.Err()
}

type wrapper struct {
	clusterClient *clientcfg.Cluster
	spaceClient   client.SpaceOperations
}

func (w wrapper) FormatTarget(obj interface{}) (string, error) {
	if v, ok := obj.(*client.PublicDomain); ok {
		w.clusterClient.DefaultWorkspaceName = v.SpaceName
		workspace, err := w.clusterClient.Workspace()
		if err != nil {
			return "", nil
		}
		workspaceClient, err := workspace.Client()
		if err != nil {
			return "", nil
		}
		spaces, err := w.spaceClient.List(&types.ListOpts{})
		if err != nil {
			return "", nil
		}
		for _, s := range spaces.Data {
			if s.Name == v.SpaceName {
				ns := namespace.StackNamespace(s.ID, v.StackName)
				svc, err := workspaceClient.Service.ByID(fmt.Sprintf("%s:%s", ns, v.ServiceName))
				if err != nil {
					return "", err
				}
				if len(svc.Endpoints) == 0 {
					return "", nil
				}
				target := svc.Endpoints[0].URL
				if v.RequestTLSCert {
					target = strings.Replace(target, "http://", "https://", 1)
				}
				return target, nil
			}
		}
	}
	return "", nil
}
