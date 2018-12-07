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
	projectv1client "github.com/rancher/rio/types/client/project/v1"
	riov1client "github.com/rancher/rio/types/client/rio/v1"
	"github.com/urfave/cli"
)

type Ls struct {
}

func (l *Ls) Customize(cmd *cli.Command) {
	cmd.Flags = append(table.WriterFlags(), cmd.Flags...)
}

func (l *Ls) Run(ctx *clicontext.CLIContext) error {
	wc, err := ctx.ProjectClient()
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

	spaces, err := spaceClient.Project.List(&types.ListOpts{})
	if err != nil {
		return nil
	}
	projectNames := make(map[string]string, 0)
	for _, s := range spaces.Data {
		projectNames[s.Name] = s.ID
	}
	w := wrapper{
		clusterClient: cluster,
		spaces:        projectNames,
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
	v, ok := obj.(*projectv1client.PublicDomain)
	if !ok {
		return "", nil
	}
	w.clusterClient.DefaultProjectName = v.TargetProjectName
	project, err := w.clusterClient.Project()
	if err != nil {
		return "", nil
	}
	projectClient, err := project.Client()
	if err != nil {
		return "", nil
	}
	for name, id := range w.spaces {
		if name == v.TargetProjectName {
			ns := namespace.StackNamespace(id, v.TargetStackName)
			svc, err := projectClient.Service.ByID(fmt.Sprintf("%s:%s", ns, v.TargetName))
			if err == nil && len(svc.Endpoints) > 0 {
				target := strings.Replace(svc.Endpoints[0].URL, "http://", "https://", 1)
				return target, nil
			}
			route, err := projectClient.RouteSet.ByID(fmt.Sprintf("%s:%s", ns, v.TargetName))
			if err == nil {
				stack := w.stackByID[route.StackID]
				space := strings.SplitN(stack.ProjectID, "-", 2)[1]
				return fmt.Sprintf("https://%s.%s", service.HashIfNeed(route.Name, strings.SplitN(ns, "-", 2)[0], space), w.domain), nil
			}
		}
	}
	return "", nil
}
