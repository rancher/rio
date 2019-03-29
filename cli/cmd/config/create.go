package config

import (
	"encoding/base64"
	"fmt"
	"unicode/utf8"

	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/stack"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
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

	config := riov1.Config{}

	_, namespace, name, err := stack.ResolveSpaceStackForName(ctx, name)
	if err != nil {
		return err
	}

	content, err := util.ReadFile(file)
	if err != nil {
		return err
	}

	config.Labels = c.L_Label
	if utf8.Valid(content) {
		config.Spec.Content = string(content)
	} else {
		config.Spec.Content = base64.StdEncoding.EncodeToString(content)
		config.Spec.Encoded = true
	}

	return ctx.Create(v1.NewConfig(namespace, name, config))
}
