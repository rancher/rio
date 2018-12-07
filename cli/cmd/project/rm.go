package project

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/types/client/project/v1"
)

type Rm struct{}

func (r *Rm) Run(ctx *clicontext.CLIContext) error {
	cluster, err := ctx.Cluster()
	if err != nil {
		return err
	}
	cc, err := cluster.Client()
	if err != nil {
		return err
	}
	projects, err := cluster.Projects()
	if err != nil {
		return err
	}
	m := map[string]client.Project{}
	for _, project := range projects {
		m[project.Name] = project.Project
		m[project.ID] = project.Project
	}
	for _, arg := range ctx.CLI.Args() {
		p, ok := m[arg]
		if !ok {
			continue
		}
		if err := cc.Project.Delete(&p); err != nil {
			return err
		}
	}
	return nil
}
