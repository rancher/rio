package kubectl

import (
	"os"

	"github.com/docker/docker/pkg/reexec"
	"github.com/rancher/rio/cli/server"
	"github.com/urfave/cli"
)

func NewKubectlCommand() cli.Command {
	return cli.Command{
		Name:            "kubectl",
		Usage:           "Run kubectl to troubelshoot kubernetes backend",
		Category:        "DEBUGGING",
		Hidden:          true,
		SkipFlagParsing: true,
		SkipArgReorder:  true,
		Action:          kubectl,
	}
}

func kubectl(app *cli.Context) error {
	ctx, err := server.NewContext(app)
	if err != nil {
		return err
	}
	defer ctx.Close()

	cmd := reexec.Command("kubectl")
	cmd.Args = append(cmd.Args, os.Args[2:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
