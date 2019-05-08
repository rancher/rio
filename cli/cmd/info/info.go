package info

import (
	"fmt"
	"strings"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/version"
	"github.com/urfave/cli"
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

	domain, err := ctx.Domain()
	if err != nil {
		return err
	}

	builder.WriteString(fmt.Sprintf("Rio Version: %s\n", version.Version))
	builder.WriteString(fmt.Sprintf("Cluster Domain: %s\n", domain))
	builder.WriteString(fmt.Sprintf("System Namespace: %s", ctx.SystemNamespace))
	fmt.Println(builder.String())
	return nil
}
