package setcontext

import (
	"github.com/pkg/errors"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/clientcfg"
	"github.com/urfave/cli"
)

func SetContext() cli.Command {
	return cli.Command{
		Name:      "set-context",
		Usage:     "Set default context",
		Action:    clicontext.DefaultAction(setContext),
		ArgsUsage: "[CLUSTER_NAME]",
	}
}

func setContext(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return errors.New("Incorrect Usage. Example: `rio set-context $CLUSTER_NAME`")
	}
	clusterName := ctx.CLI.Args()[0]
	clusters, err := clientcfg.ListClusters(ctx.ClusterDir())
	if err != nil {
		return err
	}
	for i, cluster := range clusters {
		if cluster.Name == clusterName {
			clusters[i].Default = true
		} else {
			clusters[i].Default = false
		}
		if err := ctx.SaveCluster(&clusters[i], false); err != nil {
			return err
		}
	}
	return nil
}
