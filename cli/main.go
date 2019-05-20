package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rancher/rio/cli/cmd/uninstall"

	"github.com/rancher/rio/cli/cmd/systemlogs"

	"github.com/docker/docker/pkg/reexec"
	"github.com/rancher/rio/cli/cmd/apply"
	"github.com/rancher/rio/cli/cmd/attach"
	"github.com/rancher/rio/cli/cmd/builds"
	"github.com/rancher/rio/cli/cmd/config"
	"github.com/rancher/rio/cli/cmd/create"
	"github.com/rancher/rio/cli/cmd/edit"
	"github.com/rancher/rio/cli/cmd/exec"
	"github.com/rancher/rio/cli/cmd/export"
	"github.com/rancher/rio/cli/cmd/externalservice"
	"github.com/rancher/rio/cli/cmd/feature"
	"github.com/rancher/rio/cli/cmd/info"
	"github.com/rancher/rio/cli/cmd/inspect"
	"github.com/rancher/rio/cli/cmd/install"
	"github.com/rancher/rio/cli/cmd/logs"
	"github.com/rancher/rio/cli/cmd/promote"
	"github.com/rancher/rio/cli/cmd/ps"
	"github.com/rancher/rio/cli/cmd/publicdomain"
	"github.com/rancher/rio/cli/cmd/render"
	"github.com/rancher/rio/cli/cmd/revision"
	"github.com/rancher/rio/cli/cmd/rm"
	"github.com/rancher/rio/cli/cmd/route"
	"github.com/rancher/rio/cli/cmd/run"
	"github.com/rancher/rio/cli/cmd/scale"
	"github.com/rancher/rio/cli/cmd/secrets"
	"github.com/rancher/rio/cli/cmd/stage"
	"github.com/rancher/rio/cli/cmd/tui"
	"github.com/rancher/rio/cli/cmd/weight"
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/pkg/version"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	desc = `For node scheduling arguments (--node-require, --node-require-any,
   --node-preferred) the expression is evaluated against the node labels using the
   following syntax.


      foo=bar              Node with label foo must have the value bar
      foo!=bar             If node has a label with key foo it must not equal to bar
      foo                  Node must have a label with key foo and any value
      foo in (bar, baz)    Node must have a label with key foo and a value of bar or baz
      foo notin (bar, baz) If node has a label with key foo it must not equal to bar or baz
      foo > 3              Node must have a label with key foo and a value greater than 3
      foo < 3              Node must have a label with key foo and a value less than 3
      !foo                 Node must not have a label with key foo with any value
      expr && expr         Any above expression can be combined so that all expression must be true
`
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
		cli.BoolFlag{
			Name:        "wait,w",
			Usage:       "Wait for resource to reach resting state",
			Destination: &cfg.Wait,
		},
		cli.IntFlag{
			Name:        "wait-timeout",
			Usage:       "Timeout in seconds to wait",
			Value:       600,
			Destination: &cfg.WaitTimeout,
		},
		cli.StringFlag{
			Name:        "wait-state",
			Usage:       "State to wait for (active, healthy, etc)",
			Destination: &cfg.WaitState,
		},
		cli.StringFlag{
			Name:        "system-namespace",
			Value:       "rio-system",
			Destination: &cfg.SystemNamespace,
		},
		cli.StringFlag{
			Name:        "namespace,n",
			Usage:       "Specify which namespace in kubernetes to use",
			EnvVar:      "NAMESPACE",
			Destination: &cfg.DefaultNamespace,
		},
		cli.StringFlag{
			Name:        "kubeconfig",
			Usage:       "Kubeconfig file to use",
			EnvVar:      "KUBECONFIG",
			Value:       "${HOME}/.kube/config",
			Destination: &cfg.Kubeconfig,
		},
		cli.BoolFlag{
			Name:  "system,s",
			Usage: "Only show system resources",
		},
	}

	app.Commands = []cli.Command{
		info.Info(app),
		config.Config(app),
		publicdomain.PublicDomain(app),
		externalservice.ExternalService(app),
		feature.Feature(app),
		secrets.Secrets(app),
		builds.Builds(app),

		builder.Command(&ps.Ps{},
			"List services and containers",
			appName+" ps [OPTIONS] [STACK...]",
			""),

		builder.Command(&run.Run{},
			"Create and run a new service",
			appName+" run [OPTIONS] IMAGE [COMMAND] [ARG...]",
			desc),
		builder.Command(&create.Create{},
			"Create a new service",
			appName+" create [OPTIONS] IMAGE [COMMAND] [ARG...]",
			desc),
		builder.Command(&scale.Scale{},
			"Scale a service",
			appName+" scale [SERVICE=NUMBER...]",
			""),
		builder.Command(&rm.Rm{},
			"Delete a service or stack",
			appName+" rm ID_OR_NAME",
			""),
		builder.Command(&inspect.Inspect{},
			"Print the raw API output of a resource",
			appName+" inspect [ID_OR_NAME...]",
			""),

		builder.Command(&edit.Edit{},
			"Edit a service or stack",
			appName+" edit ID_OR_NAME",
			""),
		builder.Command(&render.Render{},
			"Render rio file to k8s manifests",
			appName+" render FILENAME",
			""),
		builder.Command(&apply.Apply{},
			"Bring up a stack",
			appName+" up [OPTIONS] [[STACK_NAME] FILE|-]",
			""),
		builder.Command(&export.Export{},
			"Export a stack",
			appName+" export STACK_ID_OR_NAME",
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
			"Print logs from containers",
			appName+" logs [OPTIONS] [CONTAINER_OR_SERVICE...]",
			""),

		builder.Command(&install.Install{},
			"Install rio management plane",
			appName+" install [OPTIONS]",
			""),
		builder.Command(&uninstall.Uninstall{},
			"Uninstall rio",
			appName+" uninstall [OPTIONS]",
			""),
		builder.Command(&revision.Revision{},
			"List service revisions",
			appName+" revision [OPTIONS] [APP...]",
			""),
		builder.Command(&stage.Stage{},
			"Stage a new revision of a service",
			appName+" stage [OPTIONS] SERVICE_ID_NAME",
			""),
		builder.Command(&promote.Promote{},
			"Promote a staged version to latest",
			appName+" promote [SERVICE_ID_NAME]",
			""),
		builder.Command(&weight.Weight{},
			"Weight a percentage of traffic to a staged service",
			appName+" weight [OPTIONS] [SERVICE_REVISION=PERCENTAGE...]",
			""),
		builder.Command(&tui.Tui{},
			"Terminal interactive UI",
			appName+" tui",
			""),
		builder.Command(&systemlogs.SystemLogs{},
			"View system log for Rio management plane",
			appName+" systemlogs",
			""),
		route.Route(app),
	}
	app.Before = func(ctx *cli.Context) error {
		if err := cfg.Validate(); err != nil {
			if len(ctx.Args()) > 0 && ctx.Args()[0] != "install" && ctx.Args()[0] != "uninstall" {
				return err
			}
		}
		cc := clicontext.CLIContext{
			Config: &cfg,
			Ctx:    context.Background(),
		}
		cc.Store(ctx.App.Metadata)
		return nil
	}
	app.ExitErrHandler = func(context *cli.Context, err error) {
		if err == clicontext.ErrNoConfig {
			printConfigUsage()
		} else {
			cli.HandleExitCoder(err)
		}
	}

	err := app.Run(args)
	if err != nil {
		logrus.Fatal(err)
	}
}

func printConfigUsage() {
	fmt.Print(`
No configuration found to contact server.  If you already have a Kubernetes cluster running then you should point your Kubernetes cluster using


	export KUBECONFIG=/path/to/config


If you don't have rio installed then you should run the following command to install Rio into your current cluster.


	rio install 


If you are just looking for general "rio" CLI usage then run

    rio --help

`)
	os.Exit(1)
}
