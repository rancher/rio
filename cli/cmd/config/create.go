package config

import (
	"fmt"

	"encoding/base64"
	"unicode/utf8"

	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/server"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	"github.com/urfave/cli"
)

type Create struct {
	L_Label map[string]string `desc:"Set meta data on a config"`
}

func (c *Create) Run(app *cli.Context) error {
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

	config := &client.Config{}

	config.SpaceID, config.StackID, config.Name, err = ctx.ResolveSpaceStackName(name)
	if err != nil {
		return err
	}

	content, err := util.ReadFile(file)
	if err != nil {
		return err
	}

	config.Labels = c.L_Label
	if utf8.Valid(content) {
		config.Content = string(content)
	} else {
		config.Content = base64.StdEncoding.EncodeToString(content)
		config.Encoded = true
	}

	config, err = ctx.Client.Config.Create(config)
	if err != nil {
		return err
	}

	fmt.Println(config.ID)
	return nil
}
