//go:generate go run pkg/codegen/cleanup/main.go
//go:generate go run ./vendor/github.com/go-bindata/go-bindata/go-bindata -o ./stacks/bindata.go -ignore bindata.go -pkg stacks -modtime 1557785965 -mode 0644 ./stacks/
//go:generate go fmt stacks/bindata.go
//go:generate go run pkg/codegen/main.go

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"

	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/server"
	"github.com/rancher/rio/pkg/version"
	"github.com/rancher/wrangler/pkg/signals"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/klog"
)

var (
	debug          bool
	level          string
	kubeconfig     string
	namespace      string
	customRegistry string
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
			Name:        "debug",
			EnvVar:      "RIO_DEBUG",
			Destination: &debug,
		},
		cli.StringFlag{
			Name:        "debug-level",
			Value:       "6",
			Destination: &level,
		},
		cli.StringFlag{
			Name:        "http-listen-port",
			Usage:       "HTTP port gateway will be listening",
			EnvVar:      "HTTP_PORT",
			Value:       constants.DefaultHTTPOpenPort,
			Destination: &constants.DefaultHTTPOpenPort,
		},
		cli.StringFlag{
			Name:        "https-listen-port",
			Usage:       "HTTPS port gateway will be listening",
			EnvVar:      "HTTPS_PORT",
			Value:       constants.DefaultHTTPSOpenPort,
			Destination: &constants.DefaultHTTPSOpenPort,
		},
		cli.StringFlag{
			Name:        "install-mode",
			Usage:       "Whether to use hostPort to export servicemesh gateway",
			EnvVar:      "INSTALL_MODE",
			Value:       constants.InstallMode,
			Destination: &constants.InstallMode,
		},
		cli.StringFlag{
			Name:        "use-ipaddresses",
			Usage:       "Manually specify IP addresses to generate rdns domain",
			EnvVar:      "IP_ADDRESSES",
			Destination: &constants.UseIPAddress,
		},
		cli.StringFlag{
			Name:        "service-mesh-mode",
			Usage:       "Specify service mesh mode",
			EnvVar:      "SM_MODE",
			Value:       constants.ServiceMeshMode,
			Destination: &constants.ServiceMeshMode,
		},
		cli.StringFlag{
			Name:   "disable-features",
			Usage:  "Manually specify features to disable",
			EnvVar: "DISABLE_FEATURES",
		},
	}
	app.Action = run

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	logrus.Infof("Starting rio-controller, version: %s, git commit: %s", version.Version, version.GitCommit)
	go func() {
		logrus.Fatal(http.ListenAndServe("127.0.0.1:6061", nil))
	}()
	if debug {
		setupDebugLogging()
		logrus.SetLevel(logrus.DebugLevel)
	}

	disableFeatures := strings.Split(c.String("disable-features"), ",")
	for _, f := range disableFeatures {
		switch f {
		case "autoscaling":
			constants.DisableAutoscaling = true
		case "build":
			constants.DisableBuild = true
		case "grafana":
			constants.DisableGrafana = true
		case "istio":
			constants.DisableIstio = true
		case "kiali":
			constants.DisableKiali = true
		case "letsencrypt":
			constants.DisableLetsencrypt = true
		case "mixer":
			constants.DisableMixer = true
		case "prometheus":
			constants.DisablePrometheus = true
		case "rdns":
			constants.DisableRdns = true
		}
	}

	ctx := signals.SetupSignalHandler(context.Background())
	if err := server.Startup(ctx, namespace, kubeconfig); err != nil {
		return err
	}

	return nil
}

func setupDebugLogging() {
	klog.InitFlags(flag.CommandLine)
	flag.CommandLine.Lookup("v").Value.Set(level)
	flag.CommandLine.Lookup("alsologtostderr").Value.Set("true")
}
