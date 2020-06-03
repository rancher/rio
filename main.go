//go:generate go run pkg/codegen/cleanup/main.go
//go:generate go run ./vendor/github.com/go-bindata/go-bindata/go-bindata -tags static -o ./stacks/bindata.go -ignore bindata.go -pkg stacks -modtime 1557785965 -mode 0644 ./stacks/
//go:generate go fmt stacks/bindata.go
//go:generate go run pkg/codegen/main.go

package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/rancher/rio/pkg/config"

	"github.com/rancher/norman/pkg/debug"
	"github.com/rancher/rio/pkg/server"
	"github.com/rancher/rio/pkg/version"
	"github.com/rancher/wrangler/pkg/signals"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	debugConfig debug.Config
	kubeconfig  string
	namespace   string
)

func main() {
	app := cli.NewApp()
	app.Name = "rio-controller"
	app.Version = fmt.Sprintf("%s (%s)", version.Version, version.GitCommit)
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "kubeconfig",
			EnvVar:      "KUBECONFIG",
			Destination: &kubeconfig,
		},
		cli.StringFlag{
			Name:        "namespace",
			EnvVar:      "RIO_NAMESPACE",
			Value:       "rio-system",
			Destination: &namespace,
		},
		cli.BoolFlag{
			Name:        "run-api-validator",
			Usage:       "Whether to run api validator webhook",
			EnvVar:      "RUN_API_VALIDATOR",
			Destination: &config.ConfigController.RunAPIValidatorWebhook,
		},
		cli.StringFlag{
			Name:        "webhook-port",
			Usage:       "Specify which port webhook should listen on",
			EnvVar:      "RUN_API_VALIDATOR_PORT",
			Destination: &config.ConfigController.WebhookPort,
		},
		cli.StringFlag{
			Name:        "webhook-host",
			Usage:       "Specify which host webhook should listen on",
			EnvVar:      "RUN_API_VALIDATOR_HOST",
			Destination: &config.ConfigController.WebhookHost,
		},
		cli.StringFlag{
			Name:        "ip-address",
			Usage:       "Specify which ip address RDNS should generate record for",
			Destination: &config.ConfigController.IPAddresses,
		},
		cli.StringFlag{
			Name:        "features",
			Usage:       "Specify which feature must be turned on",
			Destination: &config.ConfigController.Features,
			Value:       "autoscaling,build",
		},
		cli.StringFlag{
			Name:        "mesh-mode",
			Usage:       "Specify which mesh-mode to use(linkerd/istio)",
			Destination: &config.ConfigController.MeshMode,
			Value:       "istio",
		},
		cli.StringFlag{
			Name:        "gateway-service-name",
			Usage:       "Specify which external gateway service name",
			Destination: &config.ConfigController.Gateway.ServiceName,
			Value:       "",
		},
		cli.StringFlag{
			Name:        "gateway-service-namespace",
			Usage:       "Specify which external gateway service namespace",
			Destination: &config.ConfigController.Gateway.ServiceNamespace,
			Value:       "",
		},
	}
	app.Flags = append(app.Flags, debug.Flags(&debugConfig)...)
	app.Action = run

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	debugConfig.MustSetupDebug()
	logrus.Infof("Starting rio-controller, version: %s, git commit: %s", version.Version, version.GitCommit)
	go func() {
		err := http.ListenAndServe("127.0.0.1:6061", nil)
		if err != nil {
			logrus.Errorf("Failed to launch pprof on port 6061: %v", err)
		}
	}()

	ctx := signals.SetupSignalHandler(context.Background())
	if err := server.Startup(ctx, namespace, kubeconfig); err != nil {
		return err
	}

	return nil
}
