package publicdomain

import (
	"fmt"
	"strings"

	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/clientcfg"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/pkg/namespace"
	projectv1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"github.com/urfave/cli"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Ls struct {
}

func (l *Ls) Customize(cmd *cli.Command) {
	cmd.Flags = append(table.WriterFlags(), cmd.Flags...)
}

func (l *Ls) Run(ctx *clicontext.CLIContext) error {
	cluster, err := ctx.Cluster()
	if err != nil {
		return err
	}
	domain, err := cluster.Domain()
	if err != nil {
		return err
	}
	client, err := cluster.KubeClient()
	if err != nil {
		return err
	}
	publicDomains, err := client.Project.PublicDomains("").List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	stackByID, err := util.StacksByID(client, "")
	if err != nil {
		return err
	}
	writer := table.NewWriter([][]string{
		{"DOMAIN", "{{.DomainName}}"},
		{"TARGET", "{{. | formatTarget}}"},
	}, ctx)
	defer writer.Close()

	projects, err := client.Core.Namespaces("").List(metav1.ListOptions{
		LabelSelector: "rio.cattle.io/project=true",
	})
	if err != nil {
		return nil
	}
	projectNames := make(map[string]struct{}, 0)
	for _, s := range projects.Items {
		projectNames[s.Name] = struct{}{}
	}
	w := wrapper{
		clusterClient: cluster,
		projects:      projectNames,
		ctx:           ctx,
		domain:        domain,
		stackByID:     stackByID,
	}
	writer.AddFormatFunc("formatTarget", w.FormatTarget)
	for _, publicDomain := range publicDomains.Items {
		publicDomain.Spec.DomainName = "https://" + publicDomain.Spec.DomainName
		writer.Write(&publicDomain)
	}
	return writer.Err()
}

type wrapper struct {
	ctx           *clicontext.CLIContext
	clusterClient *clientcfg.Cluster
	projects      map[string]struct{}
	domain        string
	stackByID     map[string]*riov1.Stack
}

func (w wrapper) FormatTarget(obj interface{}) (string, error) {
	v, ok := obj.(*projectv1.PublicDomain)
	if !ok {
		return "", nil
	}

	if v.Spec.TargetStackName == "" {
		v.Spec.TargetProjectName = w.clusterClient.DefaultProjectName
	}

	client, err := w.ctx.KubeClient()
	if err != nil {
		return "", nil
	}

	for name := range w.projects {
		if name == v.Spec.TargetProjectName {
			ns := namespace.StackNamespace(name, v.Spec.TargetStackName)
			svc, err := client.Rio.Services(v.Spec.TargetStackName).Get(v.Spec.TargetName, metav1.GetOptions{})
			if err == nil && len(svc.Status.Endpoints) > 0 {
				target := strings.Replace(svc.Status.Endpoints[0].URL, "http://", "https://", 1)
				return target, nil
			}
			route, err := client.Rio.RouteSets(v.Spec.TargetStackName).Get(v.Spec.TargetName, metav1.GetOptions{})
			if err == nil {
				stack := w.stackByID[route.Spec.StackName]
				return fmt.Sprintf("https://%s.%s", namespace.HashIfNeed(route.Name, strings.SplitN(ns, "-", 2)[0], stack.Namespace), w.domain), nil
			}
		}
	}
	return "", nil
}
