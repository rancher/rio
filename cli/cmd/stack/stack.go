package stack

import (
	"fmt"

	"github.com/rancher/rio/cli/pkg/clicontext"

	"github.com/rancher/norman/clientbase"
	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/cli/pkg/waiter"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	"github.com/urfave/cli"
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
	Stack client.Stack
}

func stackLs(ctx *clicontext.CLIContext) error {
	wc, err := ctx.WorkspaceClient()
	if err != nil {
		return err
	}

	collection, err := wc.Stack.List(util.DefaultListOpts())
	if err != nil {
		return err
	}

	writer := table.NewWriter([][]string{
		{"NAME", "Stack.Name"},
		{"STATE", "Stack.State"},
		{"CREATED", "{{.Stack.Created | ago}}"},
		{"DESC", "Stack.Description"},
		{"DETAIL", "Stack.TransitioningMessage"},
	}, ctx)

	defer writer.Close()

	for _, item := range collection.Data {
		writer.Write(&Data{
			ID:    item.ID,
			Stack: item,
		})
	}

	return writer.Err()
}

func stackCreate(ctx *clicontext.CLIContext) error {
	names := []string{""}
	if len(ctx.CLI.Args()) > 0 {
		names = ctx.CLI.Args()
	}

	wc, err := ctx.WorkspaceClient()
	if err != nil {
		return err
	}

	w, err := waiter.NewWaiter(ctx)
	if err != nil {
		return err
	}

	var lastErr error
	for _, name := range names {
		stack := &client.Stack{
			Name:        name,
			Description: ctx.CLI.String("description"),
		}

		stack, err = wc.Stack.Create(stack)
		if err != nil {
			lastErr = err
		}

		w.Add(&stack.Resource)
	}

	if lastErr != nil {
		return lastErr
	}

	return w.Wait(ctx.Ctx)
}

func stackRm(ctx *clicontext.CLIContext) error {
	names := ctx.CLI.Args()

	wc, err := ctx.WorkspaceClient()
	if err != nil {
		return err
	}

	w, err := waiter.NewWaiter(ctx)
	if err != nil {
		return err
	}

	var lastErr error
	for _, name := range names {
		stack, err := lookup.Lookup(ctx, name, client.StackType)
		if err != nil {
			return err
		}

		err = wc.Ops.DoDelete(stack.Links[clientbase.SELF])
		if err != nil {
			lastErr = err
			continue
		}

		w.Add(&stack.Resource)
	}

	if lastErr != nil {
		return lastErr
	}

	return w.Wait(ctx.Ctx)
}

func stackUpdate(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return fmt.Errorf("at least on stack name is required")
	}

	names := ctx.CLI.Args()

	wc, err := ctx.WorkspaceClient()
	if err != nil {
		return err
	}

	w, err := waiter.NewWaiter(ctx)
	if err != nil {
		return err
	}

	var lastErr error
	for _, name := range names {
		stack, err := lookup.Lookup(ctx, name, client.StackType)
		if err != nil {
			return err
		}

		resp := &client.Stack{}
		err = wc.Ops.DoUpdate(client.StackType, &stack.Resource, &client.Stack{
			Description: ctx.CLI.String("description"),
		}, resp)
		if err != nil {
			lastErr = err
			continue
		}

		w.Add(&stack.Resource)
	}

	if lastErr != nil {
		return lastErr
	}

	return w.Wait(ctx.Ctx)
}
