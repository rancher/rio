package config

import (
	"encoding/base64"
	"sort"

	"github.com/docker/go-units"
	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/cli/server"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	"github.com/urfave/cli"
)

type Data struct {
	ID     string
	Stack  *client.Stack
	Config client.Config
}

type Ls struct {
	L_Label map[string]string `desc:"Set meta data on a container"`
}

func (l *Ls) Customize(cmd *cli.Command) {
	cmd.Flags = append(cmd.Flags, table.WriterFlags()...)
}

func (l *Ls) Run(app *cli.Context) error {
	ctx, err := server.NewContext(app)
	if err != nil {
		return err
	}
	defer ctx.Close()

	configs, err := ctx.Client.Config.List(util.DefaultListOpts())
	if err != nil {
		return err
	}

	writer := table.NewWriter([][]string{
		{"NAME", "{{stackScopedName .Stack.Name .Config.Name}}"},
		{"CREATED", "{{.Config.Created | ago}}"},
		{"SIZE", "{{size .Config}}"},
	}, app)
	defer writer.Close()

	writer.AddFormatFunc("size", Base64Size)

	stackByID, err := util.StacksByID(ctx)
	if err != nil {
		return err
	}

	sort.Slice(configs.Data, func(i, j int) bool {
		return configs.Data[i].ID < configs.Data[j].ID
	})

	for i, service := range configs.Data {
		writer.Write(&Data{
			ID:     service.ID,
			Config: configs.Data[i],
			Stack:  stackByID[service.StackID],
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
