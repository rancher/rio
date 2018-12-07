package config

import (
	"encoding/base64"
	"sort"

	"github.com/docker/go-units"
	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/types/client/rio/v1"
	"github.com/urfave/cli"
)

type Data struct {
	ID     string
	Stack  *client.Stack
	Config *client.Config
}

type Ls struct {
	L_Label map[string]string `desc:"Set meta data on a container"`
}

func (l *Ls) Customize(cmd *cli.Command) {
	cmd.Flags = append(cmd.Flags, table.WriterFlags()...)
}

func (l *Ls) Run(ctx *clicontext.CLIContext) error {
	wc, err := ctx.ProjectClient()
	if err != nil {
		return err
	}

	cluster, err := ctx.Cluster()
	if err != nil {
		return err
	}

	configs, err := wc.Config.List(util.DefaultListOpts())
	if err != nil {
		return err
	}

	writer := table.NewWriter([][]string{
		{"NAME", "{{stackScopedName .Stack.Name .Config.Name}}"},
		{"CREATED", "{{.Config.Created | ago}}"},
		{"SIZE", "{{size .Config}}"},
	}, ctx)
	defer writer.Close()

	writer.AddFormatFunc("size", Base64Size)
	writer.AddFormatFunc("stackScopedName", table.FormatStackScopedName(cluster))

	stackByID, err := util.StacksByID(wc)
	if err != nil {
		return err
	}

	sort.Slice(configs.Data, func(i, j int) bool {
		return configs.Data[i].ID < configs.Data[j].ID
	})

	for i, config := range configs.Data {
		writer.Write(&Data{
			ID:     config.ID,
			Config: &configs.Data[i],
			Stack:  stackByID[config.StackID],
		})
	}

	return writer.Err()
}

func Base64Size(data interface{}) (string, error) {
	c, ok := data.(client.Config)
	if !ok {
		return "", nil
	}

	size := len(c.Content)
	if size > 0 && c.Encoded {
		content, err := base64.StdEncoding.DecodeString(c.Content)
		if err != nil {
			return "", err
		}
		size = len(content)
	}

	return units.HumanSize(float64(size)), nil
}
