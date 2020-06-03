package system

import (
	"bufio"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/rancher/rio/cli/cmd/edit/edit"
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/logger"
	"github.com/rancher/rio/cli/pkg/tables"
	"github.com/rancher/rio/pkg/config"
	"github.com/urfave/cli"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func System(app *cli.App) cli.Command {
	logs := builder.Command(&Logs{},
		"View system logs",
		app.Name+" system logs",
		"")
	feature := builder.Command(&Feature{},
		"View/Edit feature setting",
		app.Name+" feature [OPTIONS]",
		"")
	return cli.Command{
		Name:     "system",
		Usage:    "System settings",
		Category: "SUB COMMANDS",
		Subcommands: []cli.Command{
			logs,
			feature,
		},
	}

}

type Logs struct {
}

func (s *Logs) Run(ctx *clicontext.CLIContext) error {
	pods, err := ctx.Core.Pods(ctx.SystemNamespace).List(metav1.ListOptions{
		LabelSelector: "rio-controller=true",
	})
	if err != nil {
		return err
	}

	if len(pods.Items) == 0 {
		return fmt.Errorf("failed to find rio controller pod")
	}

	factory := logger.NewColorLoggerFactory()
	logger := factory.CreateContainerLogger("rio-controller")
	req := ctx.Core.Pods(pods.Items[0].Namespace).GetLogs(pods.Items[0].Name, &v1.PodLogOptions{
		Follow: true,
	})
	reader, err := req.Stream()
	if err != nil {
		return err
	}
	defer reader.Close()

	sc := bufio.NewScanner(reader)
	for sc.Scan() {
		logger.Out(append(sc.Bytes(), []byte("\n")...))
	}

	return nil
}

type Feature struct {
	Edit bool `desc:"edit system configuration"`
}

type featureData struct {
	Name        string
	Enabled     bool
	Description string
}

func (s *Feature) Run(ctx *clicontext.CLIContext) error {
	cm, err := ctx.Core.ConfigMaps(ctx.SystemNamespace).Get(config.ConfigName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	conf, err := config.FromConfigMap(cm)
	if err != nil {
		return err
	}

	if s.Edit {
		data, err := json.MarshalIndent(conf, "", "  ")
		if err != nil {
			return err
		}

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

	var data []featureData
	for name, feature := range conf.Features {
		data = append(data, featureData{
			Name:        name,
			Enabled:     *feature.Enabled,
			Description: feature.Description,
		})
	}
	sort.Slice(data, func(i, j int) bool {
		return data[i].Name < data[j].Name
	})
	writer := tables.NewFeatures(ctx)
	defer writer.TableWriter().Close()
	for _, obj := range data {
		writer.TableWriter().Write(obj)
	}
	return nil
}
