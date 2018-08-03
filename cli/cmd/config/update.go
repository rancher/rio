package config

import (
	"fmt"

	"encoding/base64"
	"unicode/utf8"

	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/server"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	"github.com/urfave/cli"
)

type Update struct {
	L_Label map[string]string `desc:"Set meta data on a config"`
}

func (c *Update) Run(app *cli.Context) error {
	ctx, err := server.NewContext(app)
	if err != nil {
		return err
	}
	defer ctx.Close()

	if len(app.Args()) != 2 {
		return fmt.Errorf("two arguments are required")
	}

	name := app.Args()[0]
	file := app.Args()[1]

	resource, err := lookup.Lookup(ctx.Client, name, client.ConfigType)
	if err != nil {
		return err
	}

	content, err := util.ReadFile(file)
	if err != nil {
		return err
	}

	err = RunUpdate(ctx, resource.ID, content, c.L_Label)
	if err == nil {
		fmt.Println(resource.ID)
	}
	return err
}

func RunUpdate(ctx *server.Context, id string, content []byte, labels map[string]string) error {
	config, err := ctx.Client.Config.ByID(id)
	if err != nil {
		return err
	}

	if len(labels) > 0 {
		config.Labels = labels
	}
	if utf8.Valid(content) {
		config.Content = string(content)
		config.Encoded = false
	} else {
		config.Content = base64.StdEncoding.EncodeToString(content)
		config.Encoded = true
	}

	_, err = ctx.Client.Config.Update(config, config)
	return err
}
