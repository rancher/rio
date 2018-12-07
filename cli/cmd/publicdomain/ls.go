package publicdomain

import (
	"fmt"
	"strings"

	"github.com/rancher/norman/types"
	"github.com/rancher/rio/api/service"
	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/clientcfg"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/pkg/namespace"
	riov1client "github.com/rancher/rio/types/client/rio/v1beta1"
	"github.com/rancher/rio/types/client/space/v1beta1"
	"github.com/urfave/cli"
)

type Ls struct {
}

func (l *Ls) Customize(cmd *cli.Command) {
	cmd.Flags = append(table.WriterFlags(), cmd.Flags...)
}

func (l *Ls) Run(ctx *clicontext.CLIContext) error {
	wc, err := ctx.WorkspaceClient()
	if err != nil {
		return err
	}
	cluster, err := ctx.Cluster()
	if err != nil {
		return err
	}
	domain, err := cluster.Domain()
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
	stackByID, err := util.StacksByID(wc)
	if err != nil {
		return err
	}
	writer := table.NewWriter([][]string{
		{"DOMAIN", "{{.DomainName}}"},
		{"TARGET", "{{. | formatTarget}}"},
	}, ctx)
	defer writer.Close()

	spaces, err := spaceClient.Space.List(&types.ListOpts{})
	if err != nil {
		return nil
	}
	spaceNames := make(map[string]string, 0)
	for _, s := range spaces.Data {
		spaceNames[s.Name] = s.ID
	}
	w := wrapper{
		clusterClient: cluster,
		spaces:        spaceNames,
		ctx:           ctx,
		domain:        domain,
		stackByID:     stackByID,
	}
	writer.AddFormatFunc("formatTarget", w.FormatTarget)
	for _, publicDomain := range publicDomains.Data {
		publicDomain.DomainName = "https://" + publicDomain.DomainName
		writer.Write(&publicDomain)
	}
	return writer.Err()
}

type wrapper struct {
	ctx           *clicontext.CLIContext
	clusterClient *clientcfg.Cluster
	spaces        map[string]string
	domain        string
	stackByID     map[string]*riov1client.Stack
}

func (w wrapper) FormatTarget(obj interface{}) (string, error) {
	v, ok := obj.(*client.PublicDomain)
	if !ok {
		return "", nil
	}
	w.clusterClient.DefaultWorkspaceName = v.TargetWorkspaceName
	workspace, err := w.clusterClient.Workspace()
	if err != nil {
		return "", nil
	}
	workspaceClient, err := workspace.Client()
	if err != nil {
		return "", nil
	}
	for name, id := range w.spaces {
		if name == v.TargetWorkspaceName {
			ns := namespace.StackNamespace(id, v.TargetStackName)
			svc, err := workspaceClient.Service.ByID(fmt.Sprintf("%s:%s", ns, v.TargetName))
			if err == nil && len(svc.Endpoints) > 0 {
				target := strings.Replace(svc.Endpoints[0].URL, "http://", "https://", 1)
				return target, nil
			}
			route, err := workspaceClient.RouteSet.ByID(fmt.Sprintf("%s:%s", ns, v.TargetName))
			if err == nil {
				stack := w.stackByID[route.StackID]
				space := strings.SplitN(stack.SpaceID, "-", 2)[1]
				return fmt.Sprintf("https://%s.%s", service.HashIfNeed(route.Name, strings.SplitN(ns, "-", 2)[0], space), w.domain), nil
			}
		}
	}
	return "", nil
}
