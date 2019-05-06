//go:generate go run pkg/codegen/cleanup/main.go
//go:generate go run vendor/github.com/jteeuwen/go-bindata/go-bindata/AppendSliceValue.go vendor/github.com/jteeuwen/go-bindata/go-bindata/main.go vendor/github.com/jteeuwen/go-bindata/go-bindata/version.go -o ./stacks/bindata.go -ignore bindata.go -pkg stacks ./stacks/
//go:generate go fmt stacks/bindata.go
//go:generate go run pkg/codegen/main.go

package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"strings"

	_ "net/http/pprof"

	"github.com/rancher/rio/pkg/server"
	"github.com/rancher/rio/version"
	"github.com/rancher/wrangler/pkg/signals"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"k8s.io/klog"
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
	app.Version = version.Version
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "kubeconfig",
			EnvVar:      "KUBECONFIG",
			Destination: &kubeconfig,
		},
		cli.StringFlag{
			Name:        "namespace",
			EnvVar:      "RIO_NAMESPACE",
			Destination: &namespace,
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
	go func() {
		logrus.Fatal(http.ListenAndServe("localhost:6061", nil))
	}()
	if debug {
		setupDebugLogging()
		logrus.SetLevel(logrus.DebugLevel)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	kubeconfig = strings.Replace(kubeconfig, "${HOME}", homeDir, -1)
	kubeconfig = strings.Replace(kubeconfig, "$HOME", homeDir, -1)

	if os.Getenv("RIO_IN_CLUSTER") != "" {
		kubeconfig = ""
	}

	ctx := signals.SetupSignalHandler(context.Background())
	if err := server.Startup(ctx, namespace, customRegistry, kubeconfig); err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}

func setupDebugLogging() {
	flag.Set("alsologtostderr", "true")
	flag.Parse()

	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)
}
