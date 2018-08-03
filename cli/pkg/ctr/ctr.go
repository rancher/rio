package ctr

import (
	"fmt"
	"os"

	"github.com/containerd/containerd/cmd/ctr/app"
	"github.com/docker/docker/pkg/reexec"
	"github.com/urfave/cli"
)

var pluginCmds []cli.Command

func init() {
	reexec.Register("ctr", main)
}

func main() {
	app := app.New()
	app.Commands = append(app.Commands, pluginCmds...)
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "ctr: %s\n", err)
		os.Exit(1)
	}
}
