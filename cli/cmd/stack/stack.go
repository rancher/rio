package stack

import (
	"fmt"
	"io"
	"os"

	"github.com/rancher/rio/cli/pkg/clientcfg"

	"github.com/rancher/rio/cli/pkg/mapper"

	"github.com/rancher/rio/cli/pkg/types"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/table"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"github.com/urfave/cli"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Stack() cli.Command {
	return cli.Command{
		Name:      "stacks",
		ShortName: "stack",
		Usage:     "Operations on stacks",
		Action:    clicontext.DefaultAction(stackLs),
		Flags:     table.WriterFlags(),
		Category:  "SUB COMMANDS",
		Subcommands: []cli.Command{
			{
				Name:      "ls",
				Usage:     "List stacks",
				ArgsUsage: "None",
				Action:    clicontext.Wrap(stackLs),
				Flags:     table.WriterFlags(),
			},
			{
				Name:      "create",
				Usage:     "Create a stack",
				ArgsUsage: "None",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "d,description",
						Usage: "Description for stack",
					},
				},
				Action: clicontext.Wrap(stackCreate),
			},
			{
				Name:      "delete",
				ShortName: "rm",
				Usage:     "Delete a stack",
				ArgsUsage: "None",
				Action:    clicontext.Wrap(stackRm),
			},
			{
				Name:      "update",
				Usage:     "Update a stack",
				ArgsUsage: "None",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "d,description",
						Usage: "Description for stack",
					},
				},
				Action: clicontext.Wrap(stackUpdate),
			},
		},
	}
}

type Data struct {
	ID    string
	Stack riov1.Stack
}

func stackLs(ctx *clicontext.CLIContext) error {
	project, err := ctx.Project()
	if err != nil {
		return err
	}

	client, err := ctx.KubeClient()
	if err != nil {
		return err
	}

	data, err := ListStacks(client, project.Project.Name)
	if err != nil {
		return err
	}

	writer := NewWriter(ctx, os.Stdout)
	defer writer.Close()

	for _, item := range data {
		writer.Write(item)
	}

	return writer.Err()
}

func stackCreate(ctx *clicontext.CLIContext) error {
	names := []string{""}
	if len(ctx.CLI.Args()) > 0 {
		names = ctx.CLI.Args()
	}

	project, err := ctx.Project()
	if err != nil {
		return err
	}

	client, err := ctx.KubeClient()
	if err != nil {
		return err
	}

	var lastErr error
	for _, name := range names {
		stack := &riov1.Stack{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: riov1.StackSpec{
				Description: ctx.CLI.String("description"),
			},
		}

		stack, err = client.Rio.Stacks(project.Project.Name).Create(stack)
		if err != nil {
			lastErr = err
		}
	}

	if lastErr != nil {
		return lastErr
	}

	return nil
}

func stackRm(ctx *clicontext.CLIContext) error {
	names := ctx.CLI.Args()

	project, err := ctx.Project()
	if err != nil {
		return err
	}

	client, err := ctx.KubeClient()
	if err != nil {
		return err
	}

	var lastErr error
	for _, name := range names {
		resource, err := lookup.Lookup(ctx, name, types.StackType)
		if err != nil {
			return err
		}

		if err := client.Rio.Stacks(project.Project.Name).Delete(resource.Name, &metav1.DeleteOptions{}); err != nil {
			lastErr = err
		}
	}

	if lastErr != nil {
		return lastErr
	}

	return nil
}

func stackUpdate(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return fmt.Errorf("at least on stack name is required")
	}

	names := ctx.CLI.Args()

	project, err := ctx.Project()
	if err != nil {
		return err
	}

	client, err := ctx.KubeClient()
	if err != nil {
		return err
	}

	var lastErr error
	for _, name := range names {
		resource, err := lookup.Lookup(ctx, name, types.StackType)
		if err != nil {
			return err
		}

		stack, err := client.Rio.Stacks(project.Project.Name).Get(resource.Name, metav1.GetOptions{})
		if err != nil {
			lastErr = err
			continue
		}
		stack.Spec.Description = ctx.CLI.String("description")

		if _, err := client.Rio.Stacks(project.Project.Name).Update(stack); err != nil {
			lastErr = err
			continue
		}
	}

	if lastErr != nil {
		return lastErr
	}

	return nil
}

func ListStacks(client *clientcfg.KubeClient, project string) ([]Data, error) {
	stacks, err := client.Rio.Stacks(project).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var data []Data
	for _, s := range stacks.Items {
		data = append(data, Data{
			ID:    s.Name,
			Stack: s,
		})
	}
	return data, nil
}

func NewWriter(ctx *clicontext.CLIContext, w io.Writer) *table.Writer {
	writer := table.NewWriter([][]string{
		{"NAME", "Stack.Name"},
		{"STATE", "Stack | toJson | state"},
		{"CREATED", "{{.Stack.CreationTimestamp | ago}}"},
		{"DESC", "Stack.Spec.Description"},
		{"DETAIL", "Stack | toJson | transitioning"},
	}, ctx, w)

	m := mapper.GenericStatusMapper
	writer.AddFormatFunc("state", m.FormatState)
	writer.AddFormatFunc("transitioning", m.FormatTransitionMessage)
	return writer
}
