//go:generate go run pkg/codegen/cleanup/main.go
//go:generate go run vendor/github.com/jteeuwen/go-bindata/go-bindata/AppendSliceValue.go vendor/github.com/jteeuwen/go-bindata/go-bindata/main.go vendor/github.com/jteeuwen/go-bindata/go-bindata/version.go -o ./stacks/bindata.go -ignore bindata.go -pkg stacks ./stacks/
//go:generate go fmt stacks/bindata.go
//go:generate go run pkg/codegen/main.go

package main

import (
	"context"
	"os"
	"strings"

	"github.com/docker/libcompose/version"
	"github.com/rancher/rio/pkg/server"
	"github.com/rancher/wrangler/pkg/signals"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	debug          bool
	kubeconfig     string
	namespace      string
	customRegistry string
)

func main() {
	app := cli.NewApp()
	app.Name = "rio-controller"
	app.Version = version.VERSION
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "kubeconfig",
			EnvVar:      "KUBECONFIG",
			Value:       "${HOME}/.kube/config",
			Destination: &kubeconfig,
		},
		cli.StringFlag{
			Name:        "namespace",
			EnvVar:      "RIO_NAMESPACE",
			Value:       "rio-system",
			Destination: &namespace,
		},
		cli.StringFlag{
			Name:        "custom-registry",
			EnvVar:      "CUSTOM_REGISTRY",
			Destination: &customRegistry,
		},
		cli.BoolFlag{
			Name:        "debug",
			EnvVar:      "RIO_DEBUG",
			Destination: &debug,
		},
	}
	app.Action = run

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	kubeconfig = strings.Replace(kubeconfig, "${HOME}", homeDir, -1)
	kubeconfig = strings.Replace(kubeconfig, "$HOME", homeDir, -1)

	ctx := signals.SetupSignalHandler(context.Background())
	if err := server.Startup(ctx, namespace, customRegistry, kubeconfig); err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}
