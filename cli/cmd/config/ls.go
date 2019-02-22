package config

import (
	"encoding/base64"
	"sort"

	units "github.com/docker/go-units"
	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"github.com/urfave/cli"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Data struct {
	ID     string
	Stack  *riov1.Stack
	Config *riov1.Config
}

type Ls struct {
	L_Label map[string]string `desc:"Set meta data on a container"`
}

func (l *Ls) Customize(cmd *cli.Command) {
	cmd.Flags = append(cmd.Flags, table.WriterFlags()...)
}

func (l *Ls) Run(ctx *clicontext.CLIContext) error {
	project, err := ctx.Project()
	if err != nil {
		return err
	}

	cluster, err := ctx.Cluster()
	if err != nil {
		return err
	}

	client, err := ctx.KubeClient()
	if err != nil {
		return err
	}

	configs, err := client.Rio.Configs("").List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	var confs []riov1.Config
	for _, config := range configs.Items {
		if config.Spec.ProjectName == project.Project.Name {
			confs = append(confs, config)
		}
	}

	writer := table.NewWriter([][]string{
		{"NAME", "{{stackScopedName .Stack.Name .Config.Name}}"},
		{"CREATED", "{{.Config.CreationTimestamp | ago}}"},
		{"SIZE", "{{size .Config}}"},
	}, ctx)
	defer writer.Close()

	writer.AddFormatFunc("size", Base64Size)
	writer.AddFormatFunc("stackScopedName", table.FormatStackScopedName(cluster))

	stackByID, err := util.StacksByID(client, project.Project.Name)
	if err != nil {
		return err
	}

	sort.Slice(confs, func(i, j int) bool {
		return confs[i].Name < confs[j].Name
	})

	for i, config := range confs {
		stack := stackByID[config.Spec.StackName]
		if stack == nil {
			continue
		}
		writer.Write(&Data{
			ID:     config.Name,
			Config: &confs[i],
			Stack:  stackByID[config.Spec.StackName],
		})
	}

	return writer.Err()
}

func Base64Size(data interface{}) (string, error) {
	c, ok := data.(riov1.Config)
	if !ok {
		return "", nil
	}

	size := len(c.Spec.Content)
	if size > 0 && c.Spec.Encoded {
		content, err := base64.StdEncoding.DecodeString(c.Spec.Content)
		if err != nil {
			return "", err
		}
		size = len(content)
	}

	return units.HumanSize(float64(size)), nil
}
