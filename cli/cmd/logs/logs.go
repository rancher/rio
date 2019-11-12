package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"text/template"
	"time"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	"github.com/wercker/stern/stern"
	"k8s.io/apimachinery/pkg/labels"
)

type Logs struct {
	S_Since      string `desc:"Logs since a certain time, either duration (5s, 2m, 3h) or RFC3339" default:"24h"`
	T_Timestamps bool   `desc:"Print the logs with timestamp" default:"false"`
	N_Tail       int    `desc:"Number of recent lines to print, -1 for all" default:"200"`
	C_Container  string `desc:"Print the logs of a specific container, use -a for system containers"`
	P_Previous   bool   `desc:"Print the logs for the previous instance of the container in a pod if it exists, excludes running"`
	A_All        bool   `desc:"Include hidden or systems logs when logging" default:"false"`
	NC_NoColor   bool   `desc:"Dont show color when logging" default:"false"`
	O_Output     string `desc:"Output format: [default, raw, json]"`
}

// This is based on both wercker/stern and linkerd/stern implementations

func (l *Logs) Run(ctx *clicontext.CLIContext) error {
	conf, err := l.setupConfig(ctx)
	if err != nil {
		return err
	}
	return l.Output(ctx, conf)
}

func (l *Logs) setupConfig(ctx *clicontext.CLIContext) (*stern.Config, error) {
	var err error
	config := &stern.Config{
		LabelSelector: labels.Everything(),
		Timestamps:    l.T_Timestamps,
		Namespace:     ctx.GetSetNamespace(),
	}

	if len(ctx.CLI.Args()) > 0 {
		objName := ctx.CLI.Args().First()
		obj, err := ctx.ByID(objName)
		if err != nil {
			return nil, err
		}
		if obj.Object == nil {
			return nil, errors.New("No object found")
		}
		config.Namespace = obj.Namespace
		if obj.Type == clitypes.BuildType {
			l.P_Previous = true
		}
		podName, sel, err := util.ToPodNameOrSelector(obj.Object)
		if err != nil {
			return nil, err
		}
		if podName == "" {
			config.LabelSelector = sel
			config.PodQuery, err = regexp.Compile("")
		} else {
			config.PodQuery, err = regexp.Compile(regexp.QuoteMeta(podName))
		}
		if err != nil {
			return nil, err
		}
	} else {
		config.PodQuery, _ = regexp.Compile("") // grab everything
	}

	config.ContainerQuery, err = regexp.Compile(l.C_Container)
	if err != nil {
		return nil, err
	}

	config.ExcludeContainerQuery = nil
	if l.A_All == false {
		excludeContainer, err := regexp.Compile("linkerd-proxy|linkerd-init")
		if err != nil {
			return nil, errors.Wrap(err, "failed to compile regular expression for exclude container query")
		}
		config.ExcludeContainerQuery = excludeContainer
	}

	config.Template, err = l.logFormat()
	if err != nil {
		return nil, err
	}

	tail := int64(l.N_Tail)
	config.TailLines = &tail

	config.Since, err = time.ParseDuration(l.S_Since)
	if err != nil {
		return nil, err
	}

	config.ContainerState = []string{stern.RUNNING, stern.WAITING}
	if l.P_Previous {
		config.ContainerState = []string{stern.TERMINATED}
	}

	return config, nil
}

func (l *Logs) Output(ctx *clicontext.CLIContext, conf *stern.Config) error {
	sigCh := make(chan os.Signal, 1)
	logCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	podInterface := ctx.Core.Pods(conf.Namespace)
	tails := make(map[string]*stern.Tail)

	// See: https://github.com/linkerd/linkerd2/blob/c5a85e587c143d31f814d807e0e39cb4ad5e3572/cli/cmd/logs.go#L223-L227
	logC := make(chan string, 1024)
	go func() {
		for {
			select {
			case str := <-logC:
				fmt.Fprintf(os.Stdout, str)
			case <-logCtx.Done():
				break
			}
		}
	}()

	added, _, err := stern.Watch(
		logCtx,
		podInterface,
		conf.PodQuery,
		conf.ContainerQuery,
		conf.ExcludeContainerQuery,
		conf.ContainerState,
		conf.LabelSelector,
	)
	if err != nil {
		return err
	}

	go func() {
		for a := range added {
			tailOpts := &stern.TailOptions{
				SinceSeconds: int64(conf.Since.Seconds()),
				Timestamps:   conf.Timestamps,
				TailLines:    conf.TailLines,
				Exclude:      conf.Exclude,
				Include:      conf.Include,
				Namespace:    true,
			}

			newTail := stern.NewTail(a.Namespace, a.Pod, a.Container, conf.Template, tailOpts)
			if _, ok := tails[a.GetID()]; !ok {
				tails[a.GetID()] = newTail
			}
			newTail.Start(logCtx, podInterface, logC)
		}
	}()

	<-sigCh
	return nil
}

// logFormat is based on both wercker/stern and linkerd/stern templating
func (l *Logs) logFormat() (*template.Template, error) {
	var tpl string
	switch l.O_Output {
	case "json":
		tpl = "{{json .}}\n"
	case "raw":
		tpl = "{{.Message}}"
	default:
		tpl = "{{color .PodColor .PodName}} {{color .ContainerColor .ContainerName}} {{.Message}}"
		if l.NC_NoColor {
			tpl = "{{.PodName}} {{.ContainerName}} {{.Message}}"
		}
	}
	funk := map[string]interface{}{
		"json": func(in interface{}) (string, error) {
			b, err := json.Marshal(in)
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
		"color": func(color color.Color, text string) string {
			return color.SprintFunc()(text)
		},
	}
	template, err := template.New("log").Funcs(funk).Parse(tpl)
	if err != nil {
		return nil, err
	}
	return template, nil
}
