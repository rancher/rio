package project

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Rm struct{}

func (r *Rm) Run(ctx *clicontext.CLIContext) error {
	cluster, err := ctx.Cluster()
	if err != nil {
		return err
	}
	client, err := cluster.KubeClient()
	if err != nil {
		return err
	}
	projects, err := cluster.Projects()
	if err != nil {
		return err
	}
	m := map[string]*v1.Namespace{}
	for _, project := range projects {
		m[project.Project.Name] = project.Project
	}
	for _, arg := range ctx.CLI.Args() {
		p, ok := m[arg]
		if !ok {
			continue
		}
		if err := client.Core.Namespaces("").Delete(p.Name, &metav1.DeleteOptions{}); err != nil {
			return err
		}
	}
	return nil
}
