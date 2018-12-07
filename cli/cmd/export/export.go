package export

import (
	"io"
	"os"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/output"
	"github.com/rancher/rio/cli/pkg/yamldownload"
	"github.com/rancher/rio/types/client/rio/v1"
)

var (
	exportTypes = []string{
		client.StackType,
		client.ServiceType,
	}
)

type Export struct {
	T_Type   string `desc:"Export specific type"`
	O_Output string `desc:"Output format (yaml/json)"`
}

func (e *Export) Run(ctx *clicontext.CLIContext) error {
	format, err := output.Format(e.O_Output)
	if err != nil {
		return err
	}

	cluster, err := ctx.Cluster()
	if err != nil {
		return err
	}

	args := ctx.CLI.Args()
	if len(args) == 0 {
		args = []string{cluster.DefaultStackName}
	}

	for _, arg := range args {
		types := exportTypes
		if e.T_Type != "" {
			types = []string{e.T_Type}
		}
		_, body, _, err := yamldownload.DownloadYAML(ctx, format, "export", arg, types...)
		if err != nil {
			return err
		}
		defer body.Close()

		_, err = io.Copy(os.Stdout, body)
		if err != nil {
			return err
		}
	}

	return nil
}
