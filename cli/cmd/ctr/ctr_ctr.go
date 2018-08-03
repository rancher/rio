// +build ctr

package ctr

import (
	"os"

	"github.com/docker/docker/pkg/reexec"
	"github.com/urfave/cli"

	// ctr reexec
	_ "github.com/rancher/rio/cli/pkg/ctr"
)

func ctr(app *cli.Context) error {
	cmd := reexec.Command("ctr", "-a", "/run/rio/containerd.sock", "-n", "k8s.io")
	cmd.Args = append(cmd.Args, os.Args[2:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
