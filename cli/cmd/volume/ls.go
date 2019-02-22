package volume

import (
	"sort"

	"github.com/rancher/norman/types/convert"
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
	Volume riov1.Volume
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

	client, err := ctx.KubeClient()
	if err != nil {
		return err
	}

	cluster, err := ctx.Cluster()
	if err != nil {
		return err
	}

	volumes, err := client.Rio.Volumes("").List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	var filteredVolumes []riov1.Volume
	for _, v := range volumes.Items {
		if v.Spec.ProjectName == project.Project.Name {
			filteredVolumes = append(filteredVolumes, v)
		}
	}

	writer := table.NewWriter([][]string{
		{"NAME", "{{stackScopedName .Stack.Name .Volume.Name}}"},
		{"DRIVER", "{{.Volume.Spec.Driver | driver}}"},
		{"TEMPLATE", "Volume.Spec.Template"},
		{"SIZE GB", "Volume.Spec.SizeInGB"},
		{"CREATED", "{{.Volume.CreationTimestamp | ago}}"},
	}, ctx)
	defer writer.Close()

	writer.AddFormatFunc("driver", FormatDriver)
	writer.AddFormatFunc("stackScopedName", table.FormatStackScopedName(cluster))

	stackByID, err := util.StacksByID(client, project.Project.Name)
	if err != nil {
		return err
	}

	sort.Slice(filteredVolumes, func(i, j int) bool {
		return filteredVolumes[i].Name < filteredVolumes[j].Name
	})

	for i, volume := range filteredVolumes {
		stack := stackByID[volume.Spec.StackName]
		if stack == nil {
			continue
		}
		writer.Write(&Data{
			ID:     volume.Name,
			Volume: volumes.Items[i],
			Stack:  stack,
		})
	}

	return writer.Err()
}

func FormatDriver(obj interface{}) (string, error) {
	str := convert.ToString(obj)
	if str == "" {
		return "default", nil
	}
	return str, nil
}
