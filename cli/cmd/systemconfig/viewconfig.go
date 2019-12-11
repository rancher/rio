package systemconfig

import (
	"encoding/json"
	"fmt"

	"github.com/rancher/rio/cli/cmd/edit/edit"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SystemConfig struct {
	Edit bool `desc:"edit system configuration"`
}

func (s *SystemConfig) Run(ctx *clicontext.CLIContext) error {
	cm, err := ctx.Core.ConfigMaps(ctx.SystemNamespace).Get(config.ConfigName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	conf, err := config.FromConfigMap(cm)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		return err
	}

	if s.Edit {
		update, err := edit.Loop(nil, data, func(modifiedContent []byte) error {
			if err := json.Unmarshal(modifiedContent, &conf); err != nil {
				return err
			}
			cm, err = config.SetConfig(cm, conf)
			if err != nil {
				return err
			}
			if _, err := ctx.Core.ConfigMaps(ctx.SystemNamespace).Update(cm); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}

		if !update {
			fmt.Println("No change to system config")
		}
		return nil
	}

	fmt.Println(string(data))
	return nil
}
