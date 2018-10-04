package workspace

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/clientcfg"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/urfave/cli"
)

type Data struct {
	ID        string
	Workspace *clientcfg.Workspace
}

type Ls struct {
	A_All bool `desc:"List workspaces from all clusters"`
}

func (l *Ls) Customize(cmd *cli.Command) {
	cmd.Flags = append(cmd.Flags, table.WriterFlags()...)
}

func (l *Ls) Run(ctx *clicontext.CLIContext) error {
	var (
		clusters []clientcfg.Cluster
		err      error
	)

	var writer *table.Writer
	if l.A_All {
		clusters, err = ctx.Clusters()
		if err != nil {
			return err
		}

		if len(clusters) == 1 {
			// For the purpose of listing, this cluster is default
			clusters[0].Default = true
		}

		writer = table.NewWriter([][]string{
			{"NAME", "Workspace.Name"},
			{"CLUSTER", "Workspace.Cluster.URL"},
			{"DEFAULT", "{{and .Workspace.Default .Workspace.Cluster.Default | boolToStar}}"},
		}, ctx)
	} else {
		cluster, err := ctx.Cluster()
		if err != nil {
			return err
		}
		clusters = []clientcfg.Cluster{
			*cluster,
		}

		writer = table.NewWriter([][]string{
			{"NAME", "Workspace.Name"},
			{"DEFAULT", "{{.Workspace.Default | boolToStar}}"},
		}, ctx)
	}

	writer.AddFormatFunc("boolToStar", BoolToStar)
	defer writer.Close()

	for _, cluster := range clusters {
		workspaces, err := cluster.Workspaces()
		if err != nil {
			return err
		}
		for i, workspace := range workspaces {
			writer.Write(&Data{
				ID:        workspace.ID,
				Workspace: &workspaces[i],
			})
		}
	}

	return writer.Err()
}

func BoolToStar(obj interface{}) (string, error) {
	if b, ok := obj.(bool); ok && b {
		return "*", nil
	}
	return "", nil
}
