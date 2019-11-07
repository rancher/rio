package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rancher/rio/cli/cmd/dashboard"

	"github.com/docker/docker/pkg/reexec"
	"github.com/rancher/rio/cli/cmd/attach"
	"github.com/rancher/rio/cli/cmd/builds"
	"github.com/rancher/rio/cli/cmd/config"
	"github.com/rancher/rio/cli/cmd/edit"
	"github.com/rancher/rio/cli/cmd/endpoint"
	"github.com/rancher/rio/cli/cmd/exec"
	"github.com/rancher/rio/cli/cmd/export"
	"github.com/rancher/rio/cli/cmd/externalservice"
	"github.com/rancher/rio/cli/cmd/images"
	"github.com/rancher/rio/cli/cmd/info"
	"github.com/rancher/rio/cli/cmd/inspect"
	"github.com/rancher/rio/cli/cmd/install"
	"github.com/rancher/rio/cli/cmd/kill"
	"github.com/rancher/rio/cli/cmd/linkerd"
	"github.com/rancher/rio/cli/cmd/logs"
	"github.com/rancher/rio/cli/cmd/pods"
	"github.com/rancher/rio/cli/cmd/promote"
	"github.com/rancher/rio/cli/cmd/ps"
	"github.com/rancher/rio/cli/cmd/publicdomain"
	"github.com/rancher/rio/cli/cmd/rm"
	"github.com/rancher/rio/cli/cmd/route"
	"github.com/rancher/rio/cli/cmd/run"
	"github.com/rancher/rio/cli/cmd/scale"
	"github.com/rancher/rio/cli/cmd/secrets"
	"github.com/rancher/rio/cli/cmd/stacks"
	"github.com/rancher/rio/cli/cmd/stage"
	"github.com/rancher/rio/cli/cmd/systemlogs"
	"github.com/rancher/rio/cli/cmd/uninstall"
	"github.com/rancher/rio/cli/cmd/up"
	"github.com/rancher/rio/cli/cmd/weight"
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/pkg/version"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	// all auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	// ensure gvks are loaded
	_ "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io"
	_ "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io"
	_ "github.com/rancher/wrangler-api/pkg/generated/controllers/core"
	_ "github.com/rancher/wrangler-api/pkg/generated/controllers/tekton.dev"
)

var (
	appName = filepath.Base(os.Args[0])
	cfg     = clicontext.Config{}
)

func main() {
	if reexec.Init() {
		return
	}

	args := os.Args

	app := cli.NewApp()
	app.Name = appName
	app.Usage = "Containers made simple, as they should be"
	app.Version = fmt.Sprintf("%s (%s)", version.Version, version.GitCommit)
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("%s version %s\n", app.Name, app.Version)
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "debug",
			Usage:       "Turn on debug logs",
			Destination: &cfg.Debug,
		},
		cli.StringFlag{
			Name:        "debug-level",
			Usage:       "kubernetes client-go debug level",
			Value:       "6",
			Destination: &cfg.DebugLevel,
		},
		cli.StringFlag{
			Name:        "system-namespace",
			Value:       "rio-system",
			Destination: &cfg.SystemNamespace,
		},
		cli.BoolFlag{
			Name:        "show-system,s",
			Usage:       "Show system namespace resource",
			Destination: &cfg.ShowSystemNamespace,
		},
		cli.StringFlag{
			Name:   "namespace,n",
			Usage:  "Specify which namespace in kubernetes to use",
			EnvVar: "NAMESPACE",
		},
		cli.BoolFlag{
			Name:        "all-namespaces,a,A",
			Usage:       "Whether to show all namespaces resources",
			Destination: &cfg.AllNamespace,
		},
		cli.StringFlag{
			Name:        "kubeconfig",
			Usage:       "kubeconfig file to use",
			Destination: &cfg.Kubeconfig,
		},
	}

	app.Commands = []cli.Command{
		config.Config(app),
		publicdomain.PublicDomain(app),
		externalservice.ExternalService(app),
		secrets.Secrets(app),
		builds.Builds(app),
		stacks.Stacks(app),
		endpoint.Endpoints(app),
		pods.Pods(app),

		builder.Command(&ps.Ps{},
			"List services",
			appName+" ps [OPTIONS]",
			"To view all rio services, run `rio ps`"),

		builder.Command(&images.Images{},
			"List images built from local registry",
			appName+" images",
			""),

		builder.Command(&run.Run{},
			"Create and run a new service",
			appName+" run [OPTIONS] IMAGE [COMMAND] [ARG...]",
			""),

		builder.Command(&scale.Scale{},
			"Scale a service",
			appName+" scale [SERVICE=NUMBER_OR_MIN-MAX...]",
			fmt.Sprintf("To scale services to specified scale, run `%s scale foo=5`. To enable autoscaling, run `%s scale foo=1-5`.", appName, appName)),

		builder.Command(&rm.Rm{},
			"Delete resource",
			appName+" rm [TYPE/]RESOURCE_NAME",
			""),
		builder.Command(&inspect.Inspect{},
			"Print the raw API output of a resource",
			appName+" inspect [TYPE/][NAMESPACE/]SERVICE_NAME",
			""),

		builder.Command(&edit.Edit{},
			"Edit resource",
			appName+" edit [TYPE/]RESOURCE_NAME",
			""),
		builder.Command(&export.Export{},
			"Export a namespace or service",
			appName+" export [TYPE/]NAMESPACE_OR_SERVICE",
			""),
		builder.Command(&linkerd.Linkerd{},
			"Open linkerd dashboard",
			appName+" linkerd",
			""),

		config.NewCatCommand("", app),

		builder.Command(&exec.Exec{},
			"Run a command in a running container",
			appName+" exec [OPTIONS] CONTAINER COMMAND [ARG...]",
			""),
		builder.Command(&attach.Attach{},
			"Attach to a running process in a container",
			appName+" attach [OPTIONS] CONTAINER",
			""),
		builder.Command(&logs.Logs{},
			"Print logs from services or containers",
			appName+" logs [OPTIONS] SERVICE/BUILD",
			""),
		builder.Command(&install.Install{},
			"Install rio management plane",
			appName+" install [OPTIONS]",
			""),
		builder.Command(&uninstall.Uninstall{},
			"Uninstall rio",
			appName+" uninstall [OPTIONS]",
			""),
		builder.Command(&stage.Stage{},
			"Stage a new revision of a service",
			appName+" stage [OPTIONS] SERVICE NEW_REVISION",
			""),
		builder.Command(&promote.Promote{},
			"Promote a staged version to latest and scale down other app versions",
			appName+" promote [OPTIONS] SERVICE_NAME",
			"To perform an immediate rollout, set interval and increment to 0"),
		builder.Command(&weight.Weight{},
			"Weight a service to specific weight or percentage of total app traffic",
			appName+" weight [OPTIONS] SERVICE_NAME=WEIGHT/PERCENTAGE",
			"To perform an immediate rollout, set interval and increment to 0"),
		builder.Command(&systemlogs.SystemLogs{},
			"View system log for Rio management plane",
			appName+" systemlogs",
			""),
		builder.Command(&up.Up{},
			"Apply a rio file",
			appName+" up [OPTIONS]",
			""),
		builder.Command(&dashboard.Dashboard{},
			"Open the dashboard in a browser",
			appName+" dashboard [OPTIONS]",
			""),
		builder.Command(&kill.Kill{},
			"Kill pods individually or all pods belonging to a service",
			appName+" kill [SERVICE_NAME/POD_NAME]",
			"Specify a SERVICE_NAME to kill all pods belonging to that service. Otherwise specify a POD_NAME"),
		route.Route(app),
		info.Info(app),
	}
	app.Before = func(ctx *cli.Context) error {
		if err := cfg.Validate(); err != nil {
			return err
		}
		cc := clicontext.CLIContext{
			Config: &cfg,
			Ctx:    context.Background(),
		}
		cc.Store(ctx.App.Metadata)
		return nil
	}

	err := app.Run(args)
	if err != nil {
		logrus.Fatal(err)
	}
}
