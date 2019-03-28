package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rancher/rio/cli/cmd/cow"

	"github.com/docker/docker/pkg/reexec"
	"github.com/rancher/rio/cli/cmd/attach"
	"github.com/rancher/rio/cli/cmd/cluster"
	"github.com/rancher/rio/cli/cmd/config"
	"github.com/rancher/rio/cli/cmd/create"
	"github.com/rancher/rio/cli/cmd/exec"
	"github.com/rancher/rio/cli/cmd/export"
	"github.com/rancher/rio/cli/cmd/externalservice"
	"github.com/rancher/rio/cli/cmd/feature"
	"github.com/rancher/rio/cli/cmd/inspect"
	"github.com/rancher/rio/cli/cmd/kubectl"
	"github.com/rancher/rio/cli/cmd/login"
	"github.com/rancher/rio/cli/cmd/logs"
	"github.com/rancher/rio/cli/cmd/node"
	"github.com/rancher/rio/cli/cmd/project"
	"github.com/rancher/rio/cli/cmd/promote"
	"github.com/rancher/rio/cli/cmd/ps"
	"github.com/rancher/rio/cli/cmd/publicdomain"
	"github.com/rancher/rio/cli/cmd/rm"
	"github.com/rancher/rio/cli/cmd/route"
	"github.com/rancher/rio/cli/cmd/run"
	"github.com/rancher/rio/cli/cmd/scale"
	"github.com/rancher/rio/cli/cmd/setcontext"
	"github.com/rancher/rio/cli/cmd/stack"
	"github.com/rancher/rio/cli/cmd/stage"
	"github.com/rancher/rio/cli/cmd/up"
	"github.com/rancher/rio/cli/cmd/volume"
	"github.com/rancher/rio/cli/cmd/weight"
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/clientcfg"
	_ "github.com/rancher/rio/pkg/kubectl"
	"github.com/rancher/rio/version"
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
	cfg     = clientcfg.Config{}
)

func main() {
	if reexec.Init() {
		return
	}

	flag.Set("stderrthreshold", "3")
	flag.Set("alsologtostderr", "false")
	flag.Set("logtostderr", "false")

	args := os.Args

	app := cli.NewApp()
	app.Name = appName
	app.Usage = "Containers made simple, as they should be"
	app.Version = version.Version
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
			Name:        "cluster,c",
			Usage:       "Specify which cluster to use",
			EnvVar:      "RIO_CLUSTER",
			Destination: &cfg.ClusterName,
		},
		cli.StringFlag{
			Name:        "project,p",
			Usage:       "Specify which project to use",
			EnvVar:      "RIO_PROJECT",
			Destination: &cfg.ProjectName,
		},
		cli.StringFlag{
			Name:        "config-dir",
			Value:       "${HOME}/.rancher/rio",
			Usage:       "Specify which directory to use for config",
			EnvVar:      "RIO_CONFIG_DIR",
			Destination: &cfg.Home,
		},
	}

	app.Commands = []cli.Command{
		config.Config(app),
		volume.Volume(app),
		stack.Stack(),
		project.Projects(app),
		cluster.Cluster(app),
		node.Node(),
		publicdomain.PublicDomain(app),
		externalservice.ExternalService(app),
		setcontext.SetContext(),
		feature.Feature(app),

		builder.Command(&cow.Cow{},
			"Rio Cow",
			appName+" cow",
			""),
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

		//builder.Command(&edit.Edit{},
		//	"Edit a service or stack",
		//	appName+" edit ID_OR_NAME",
		//	""),
		builder.Command(&up.Up{},
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
		route.Route(app),

		builder.Command(&login.Login{},
			"Login into Rio",
			appName+" login",
			""),

		kubectl.NewKubectlCommand(),
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
	app.ExitErrHandler = func(context *cli.Context, err error) {
		if err == clientcfg.ErrNoConfig {
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
No configuration found to contact server.  If you already have a Rio cluster running then run

    rio login

If you don't have an existing server you should run "rio server" on a Linux server.
If you are just looking for general "rio" CLI usage then run

    rio --help

`)
	os.Exit(1)
}
