package config

import (
	"encoding/base64"
	"fmt"
	"unicode/utf8"

	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	return ctx.UpdateResource(types.Resource{
		Namespace: namespace,
		Name:      name,
		Type:      types.ConfigType,
	}, func(obj runtime.Object) error {
		config := obj.(*corev1.ConfigMap)

		if len(labels) > 0 {
			config.Labels = labels
		}
		if utf8.Valid(content) {
			config.Data["content"] = string(content)
		} else {
			config.Data["content"] = base64.StdEncoding.EncodeToString(content)
		}

		return nil
	})
}
