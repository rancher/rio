package externalservice

import (
	"strings"

	"github.com/rancher/rio/cli/cmd/rm"
	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"github.com/urfave/cli"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ExternalService(app *cli.App) cli.Command {
	return cli.Command{
		Name:      "externalservices",
		Aliases:   []string{"external"},
		ShortName: "externalservice",
		Usage:     "Operation on externalservices",
		Action:    clicontext.DefaultAction(externalServiceLs),
		Flags:     table.WriterFlags(),
		Category:  "SUB COMMANDS",
		Subcommands: []cli.Command{
			{
				Name:      "ls",
				Usage:     "List external services",
				ArgsUsage: "None",
				Action:    clicontext.Wrap(externalServiceLs),
				Flags:     table.WriterFlags(),
			},
			builder.Command(&Create{},
				"Create external services",
				app.Name+" create [OPTIONS] [EXTERNAL_SERVICE] [(IP)(FQDN)(STACK/SERVICE)]",
				""),
			{
				Name:      "delete",
				ShortName: "rm",
				Usage:     "Delete a stack",
				ArgsUsage: "None",
				Action:    clicontext.Wrap(externalServiceRm),
			},
		},
	}
}

type Data struct {
	Name    string
	Target  string
	Created string
	Service *riov1.ExternalService
	Stack   *riov1.Stack
}

func externalServiceLs(ctx *clicontext.CLIContext) error {
	client, err := ctx.KubeClient()
	if err != nil {
		return err
	}

	project, err := ctx.Project()
	if err != nil {
		return err
	}

	cluster, err := ctx.Cluster()
	if err != nil {
		return err
	}
	collection, err := client.Rio.ExternalServices("").List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	var ess []riov1.ExternalService
	for _, es := range collection.Items {
		if es.Spec.ProjectName == project.Project.Name {
			ess = append(ess, es)
		}
	}

	writer := table.NewWriter([][]string{
		{"NAME", "{{stackScopedName .Stack.Name .Service.Name}}"},
		{"CREATED", "{{.Created | ago}}"},
		{"TARGET", "{{.Service.Target}}"},
	}, ctx)
	writer.AddFormatFunc("stackScopedName", table.FormatStackScopedName(cluster))
	defer writer.Close()

	stackByID, err := util.StacksByID(client, project.Project.Name)
	if err != nil {
		return err
	}

	for _, item := range ess {
		endpoint := ""
		if item.Spec.FQDN != "" {
			endpoint = item.Spec.FQDN
		} else if item.Spec.Service != "" {
			endpoint = item.Spec.Service
		} else if len(item.Spec.IPAddresses) > 0 {
			endpoint = strings.Join(item.Spec.IPAddresses, ",")
		}
		writer.Write(&Data{
			Name:    item.Name,
			Target:  endpoint,
			Created: item.CreationTimestamp.String(),
			Stack:   stackByID[item.Spec.StackName],
			Service: &item,
		})
	}

	return writer.Err()
}

func externalServiceRm(ctx *clicontext.CLIContext) error {
	return rm.Remove(ctx, clitypes.ExternalServiceType)
}
