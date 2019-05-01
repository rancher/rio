package kubectl

import (
	"os"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/urfave/cli"
)

func NewKubectlCommand() cli.Command {
	return cli.Command{
		Name:            "kubectl",
		Usage:           "Run kubectl to troubleshoot kubernetes backend",
		Category:        "DEBUGGING",
		SkipFlagParsing: true,
		SkipArgReorder:  true,
		Action:          clicontext.Wrap(kubectl),
	}
}

func kubectl(ctx *clicontext.CLIContext) error {
	return ctx.Kubectl("", "", os.Args[2:]...)
}
