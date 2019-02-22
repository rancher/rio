package config

import (
	"encoding/base64"
	"fmt"
	"unicode/utf8"

	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	resource, err := lookup.Lookup(ctx, name, types.ConfigType)
	if err != nil {
		return err
	}

	content, err := util.ReadFile(file)
	if err != nil {
		return err
	}

	err = RunUpdate(ctx, resource.Name, resource.Namespace, content, c.L_Label)
	if err == nil {
		fmt.Println(resource.Name)
	}
	return err
}

func RunUpdate(ctx *clicontext.CLIContext, name, namespace string, content []byte, labels map[string]string) error {
	client, err := ctx.KubeClient()
	if err != nil {
		return err
	}

	config, err := client.Rio.Configs(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if len(labels) > 0 {
		config.Labels = labels
	}
	if utf8.Valid(content) {
		config.Spec.Content = string(content)
		config.Spec.Encoded = false
	} else {
		config.Spec.Content = base64.StdEncoding.EncodeToString(content)
		config.Spec.Encoded = true
	}

	_, err = client.Rio.Configs(namespace).Update(config)
	return err
}
