package volume

import (
	"sort"

	"github.com/rancher/norman/types/convert"
	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	"github.com/urfave/cli"
)

type Data struct {
	ID     string
	Stack  *client.Stack
	Volume client.Volume
}

type Ls struct {
	L_Label map[string]string `desc:"Set meta data on a container"`
}

func (l *Ls) Customize(cmd *cli.Command) {
	cmd.Flags = append(cmd.Flags, table.WriterFlags()...)
}

func (l *Ls) Run(ctx *clicontext.CLIContext) error {
	wc, err := ctx.WorkspaceClient()
	if err != nil {
		return err
	}

	cluster, err := ctx.Cluster()
	if err != nil {
		return err
	}

	volumes, err := wc.Volume.List(util.DefaultListOpts())
	if err != nil {
		return err
	}

	writer := table.NewWriter([][]string{
		{"NAME", "{{stackScopedName .Stack.Name .Volume.Name}}"},
		{"DRIVER", "{{.Volume.Driver | driver}}"},
		{"TEMPLATE", "Volume.Template"},
		{"SIZE GB", "Volume.SizeInGB"},
		{"STATE", "Volume.State"},
		{"CREATED", "{{.Volume.Created | ago}}"},
		{"DETAIL", "{{first .Volume.TransitioningMessage .Stack.TransitioningMessage}}"},
	}, ctx)
	defer writer.Close()

	writer.AddFormatFunc("driver", FormatDriver)
	writer.AddFormatFunc("stackScopedName", table.FormatStackScopedName(cluster))

	stackByID, err := util.StacksByID(wc)
	if err != nil {
		return err
	}

	sort.Slice(volumes.Data, func(i, j int) bool {
		return volumes.Data[i].ID < volumes.Data[j].ID
	})

	for i, service := range volumes.Data {
		writer.Write(&Data{
			ID:     service.ID,
			Volume: volumes.Data[i],
			Stack:  stackByID[service.StackID],
		})
	}

	return writer.Err()
}

func FormatDriver(obj interface{}) (string, error) {
	str := convert.ToString(obj)
	if str == "" {
		return "default", nil
	}
	return str, nil
}
