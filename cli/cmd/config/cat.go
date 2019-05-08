package config

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/types"
	corev1 "k8s.io/api/core/v1"
)

type Cat struct {
}

func (c *Cat) Run(ctx *clicontext.CLIContext) error {
	for _, arg := range ctx.CLI.Args() {
		r, err := lookup.Lookup(ctx, arg, types.ConfigType)
		if err != nil {
			return err
		}

		config := r.Object.(*corev1.ConfigMap)

		if len(config.Data)+len(config.BinaryData) == 0 {
			continue
		}

		builder := &strings.Builder{}
		for k, v := range config.Data {
			builder.WriteString(k)
			builder.WriteString(":")
			builder.WriteString(" |- \n")
			builder.WriteString(v)
		}
		for k, v := range config.BinaryData {
			builder.WriteString(k)
			builder.WriteString(":")
			builder.WriteString(" |- \n")
			builder.WriteString(base64.StdEncoding.EncodeToString(v))
		}
		fmt.Println(builder.String())
	}

	return nil
}
