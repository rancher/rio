package config

import (
	"fmt"

	"github.com/rancher/rio/cli/pkg/stack"

	"encoding/base64"
	"unicode/utf8"

	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/types/client/rio/v1"
)

type Create struct {
	L_Label map[string]string `desc:"Set meta data on a config"`
}

func (c *Create) Run(ctx *clicontext.CLIContext) error {
	var (
		err error
	)

	if len(ctx.CLI.Args()) != 2 {
		return fmt.Errorf("two arguments are required")
	}

	name := ctx.CLI.Args()[0]
	file := ctx.CLI.Args()[1]

	wc, err := ctx.ProjectClient()
	if err != nil {
		return err
	}

	config := &client.Config{}

	config.ProjectID, config.StackID, config.Name, err = stack.ResolveSpaceStackForName(ctx, name)
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

	config, err = wc.Config.Create(config)
	if err != nil {
		return err
	}

	fmt.Println(config.ID)
	return nil
}
