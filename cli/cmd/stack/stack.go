package stack

import (
	"fmt"

	mapper2 "github.com/rancher/mapper"
	"github.com/rancher/rio/cli/cmd/rm"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/cli/pkg/tables"
	"github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/urfave/cli"
	"k8s.io/apimachinery/pkg/runtime"
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

func stackLs(ctx *clicontext.CLIContext) error {
	stacks, err := ctx.List(types.StackType)
	if err != nil {
		return err
	}

	writer := tables.NewStack(ctx)
	return writer.Write(stacks)
}

func stackCreate(ctx *clicontext.CLIContext) error {
	names := []string{""}
	if len(ctx.CLI.Args()) > 0 {
		names = ctx.CLI.Args()
	}

	var stacks []runtime.Object
	for _, name := range names {
		stacks = append(stacks, riov1.NewStack(ctx.Namespace, name, riov1.Stack{
			Spec: riov1.StackSpec{
				Description: ctx.CLI.String("description"),
			},
		}))
	}

	return ctx.MultiCreate(stacks...)
}

func stackRm(ctx *clicontext.CLIContext) error {
	return rm.Remove(ctx, types.StackType)
}

func stackUpdate(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return fmt.Errorf("at least on stack name is required")
	}

	var (
		names  = ctx.CLI.Args()
		errors []error
	)

	for _, name := range names {
		err := ctx.Update(name, types.StackType, func(obj runtime.Object) error {
			stack := obj.(*v1.Stack)
			stack.Spec.Description = ctx.CLI.String("description")
			return nil
		})
		errors = append(errors, err)
	}

	return mapper2.NewErrors(errors...)
}
