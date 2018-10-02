package cluster

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/clientcfg"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/urfave/cli"
)

type Data struct {
	ID      string
	Cluster *clientcfg.Cluster
}

type Ls struct {
}

func (l *Ls) Customize(cmd *cli.Command) {
	cmd.Flags = append(cmd.Flags, table.WriterFlags()...)
}

func (l *Ls) Run(ctx *clicontext.CLIContext) error {
	clusters, err := ctx.Clusters()
	if err != nil {
		return err
	}

	writer := table.NewWriter([][]string{
		{"ID", "Cluster.ID"},
		{"URL", "Cluster.URL"},
		{"DEFAULT", "{{.Cluster.Default | boolToStar}}"},
	}, ctx)
	defer writer.Close()

	writer.AddFormatFunc("boolToStar", BoolToStar)

	for i, cluster := range clusters {
		writer.Write(&Data{
			ID:      cluster.ID,
			Cluster: &clusters[i],
		})
	}

	return writer.Err()
}

func BoolToStar(obj interface{}) (string, error) {
	if b, ok := obj.(bool); ok && b {
		return "*", nil
	}
	return "", nil
}
