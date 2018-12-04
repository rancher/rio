package cluster

import (
	"os"

	"github.com/pkg/errors"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/clientcfg"
)

type Rm struct {
}

func (l *Rm) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return errors.New("At least one argument needs to be provided. Example: `rio cluster rm CLUSTER_NAME`")
	}
	clusters, err := clientcfg.ListClusters(ctx.ClusterDir())
	if err != nil {
		return err
	}
	m := map[string]string{}
	for _, cluster := range clusters {
		m[cluster.ID] = cluster.File
	}
	for _, arg := range ctx.CLI.Args() {
		os.RemoveAll(m[arg])
	}
	return nil
}
