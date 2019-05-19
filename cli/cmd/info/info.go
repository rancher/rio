package info

import (
	"fmt"
	"strings"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/urfave/cli"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Info(app *cli.App) cli.Command {
	return cli.Command{
		Name:   "info",
		Usage:  "Display System-Wide Information",
		Action: clicontext.DefaultAction(info),
	}
}

func info(ctx *clicontext.CLIContext) error {
	builder := &strings.Builder{}

	info, err := ctx.Project.RioInfos().Get("rio", metav1.GetOptions{})
	if err != nil {
		return err
	}

	domain, err := ctx.Domain()
	if err != nil {
		return err
	}

	builder.WriteString(fmt.Sprintf("Rio CLI Version: %s (%s)\n", info.Status.Version, info.Status.GitCommit))
	builder.WriteString(fmt.Sprintf("Cluster Domain: %s\n", domain))
	builder.WriteString(fmt.Sprintf("System Namespace: %s", info.Status.SystemNamespace))
	fmt.Println(builder.String())
	return nil
}
