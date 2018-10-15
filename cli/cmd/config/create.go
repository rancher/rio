package config

import (
	"fmt"

	"github.com/rancher/rio/cli/pkg/stack"

	"encoding/base64"
	"unicode/utf8"

	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/types/client/rio/v1beta1"
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

	wc, err := ctx.WorkspaceClient()
	if err != nil {
		return err
	}

	config := &client.Config{}

	config.SpaceID, config.StackID, config.Name, err = stack.ResolveSpaceStackForName(ctx, name)
	if err != nil {
		return err
	}

	contents, err := util.ReadFile(file)
	if err != nil {
		return err
	}
	content := []byte(contents[util.StackFileKey])

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
