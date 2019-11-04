package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/rancher/rio/cli/pkg/clicontext"
	corev1 "k8s.io/api/core/v1"
)

type Cat struct {
	K_Key []string `desc:"specify which keys to cat"`
}

func (c *Cat) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return errors.New("at least one argument is required")
	}

	for _, arg := range ctx.CLI.Args() {
		r, err := ctx.ByID(arg)
		if err != nil {
			return err
		}

		config := r.Object.(*corev1.ConfigMap)

		switch {
		case len(c.K_Key) == 0:
			if config.Data == nil {
				config.Data = map[string]string{}
			}
			for k, v := range config.BinaryData {
				config.Data[k] = base64.StdEncoding.EncodeToString(v)
			}
			if len(config.Data) > 0 {
				if err := yaml.NewEncoder(os.Stdout).Encode(config.Data); err != nil {
					return err
				}
			}
		case len(c.K_Key) == 1:
			v := config.Data[c.K_Key[0]]
			if v == "" {
				v = base64.StdEncoding.EncodeToString(config.BinaryData[c.K_Key[0]])
			}
			fmt.Println(v)
		case len(c.K_Key) > 1:
			data := map[string]string{}
			for k, v := range config.Data {
				for _, t := range c.K_Key {
					if t == k {
						data[k] = v
					}
				}
			}
			for k, v := range config.BinaryData {
				config.Data[k] = base64.StdEncoding.EncodeToString(v)
			}
			if len(data) > 0 {
				if err := yaml.NewEncoder(os.Stdout).Encode(data); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
