package config

import (
	"encoding/base64"
	"fmt"
	"unicode/utf8"

	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

type Update struct {
	L_Label map[string]string `desc:"Set meta data on a config"`
}

func (c *Update) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) != 2 {
		return fmt.Errorf("two arguments are required")
	}

	name := ctx.CLI.Args()[0]
	file := ctx.CLI.Args()[1]

	resource, err := lookup.Lookup(ctx.ClientLookup, name, client.ConfigType)
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

func RunUpdate(ctx *clicontext.CLIContext, id string, content []byte, labels map[string]string) error {
	wc, err := ctx.WorkspaceClient()
	if err != nil {
		return err
	}

	config, err := wc.Config.ByID(id)
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

	_, err = wc.Config.Update(config, config)
	return err
}
