package config

import (
	"encoding/base64"
	"fmt"
	"unicode/utf8"

	"github.com/rancher/rio/pkg/constructors"

	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/stack"
	corev1 "k8s.io/api/core/v1"
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

	namespace, name := stack.NamespaceAndName(ctx, ctx.CLI.Args()[0])
	file := ctx.CLI.Args()[1]

	config := constructors.NewConfigMap(namespace, name, corev1.ConfigMap{
		Data:       make(map[string]string),
		BinaryData: make(map[string][]byte),
	})

	content, err := util.ReadFile(file)
	if err != nil {
		return err
	}

	config.Labels = c.L_Label
	if utf8.Valid(content) {
		config.Data["content"] = string(content)
	} else {
		config.Data["content"] = base64.StdEncoding.EncodeToString(content)
	}

	return ctx.Create(config)
}
