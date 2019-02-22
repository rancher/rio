package config

import (
	"encoding/base64"
	"os"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Cat struct {
}

func (c *Cat) Run(ctx *clicontext.CLIContext) error {
	for _, arg := range ctx.CLI.Args() {
		r, err := ctx.ByID(arg, types.ConfigType)
		if err != nil {
			return err
		}
		cluster, err := ctx.Cluster()
		if err != nil {
			return err
		}
		client, err := cluster.KubeClient()
		if err != nil {
			return err
		}
		config, err := client.Rio.Configs(r.Namespace).Get(r.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if len(config.Spec.Content) == 0 {
			continue
		}

		var out []byte
		if config.Spec.Encoded {
			bytes, err := base64.StdEncoding.DecodeString(config.Spec.Content)
			if err != nil {
				return err
			}
			out = bytes
		} else {
			out = []byte(config.Spec.Content)
		}

		_, err = os.Stdout.Write(out)
		if err != nil {
			return err
		}
	}

	return nil
}
