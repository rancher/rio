package stack

import (
	"fmt"

	"github.com/rancher/norman/clientbase"
	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/cli/pkg/waiter"
	"github.com/rancher/rio/cli/server"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	"github.com/urfave/cli"
)

func Stack() cli.Command {
	return cli.Command{
		Name:      "stacks",
		ShortName: "stack",
		Usage:     "Operations on stacks",
		Action:    defaultAction(stackLs),
		Flags:     table.WriterFlags(),
		Category:  "SUB COMMANDS",
		Subcommands: []cli.Command{
			{
				Name:      "ls",
				Usage:     "List stacks",
				ArgsUsage: "None",
				Action:    stackLs,
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
				Action: stackCreate,
			},
			{
				Name:      "delete",
				ShortName: "rm",
				Usage:     "Delete a stack",
				ArgsUsage: "None",
				Action:    stackRm,
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
				Action: stackUpdate,
			},
		},
	}
}

type Data struct {
	ID    string
	Stack client.Stack
}

func stackLs(app *cli.Context) error {
	ctx, err := server.NewContext(app)
	if err != nil {
		return err
	}
	defer ctx.Close()

	collection, err := ctx.Client.Stack.List(util.DefaultListOpts())
	if err != nil {
		return err
	}

	writer := table.NewWriter([][]string{
		{"NAME", "Stack.Name"},
		{"STATE", "Stack.State"},
		{"CREATED", "{{.Stack.Created | ago}}"},
		{"DESC", "Stack.Description"},
		{"DETAIL", "Stack.TransitioningMessage"},
	}, app)

	defer writer.Close()

	for _, item := range collection.Data {
		writer.Write(&Data{
			ID:    item.ID,
			Stack: item,
		})
	}

	return writer.Err()
}

func stackCreate(app *cli.Context) error {
	ctx, err := server.NewContext(app)
	if err != nil {
		return err
	}
	defer ctx.Close()

	names := []string{""}
	if len(app.Args()) > 0 {
		names = app.Args()
	}

	w, err := waiter.NewWaiter(ctx)
	if err != nil {
		return err
	}

	var lastErr error
	for _, name := range names {
		stack := &client.Stack{
			Name:        name,
			Description: app.String("description"),
		}

		stack, err = ctx.Client.Stack.Create(stack)
		if err != nil {
			lastErr = err
		}

		w.Add(stack.ID)
	}

	if lastErr != nil {
		return lastErr
	}

	return w.Wait()
}

func stackRm(app *cli.Context) error {
	ctx, err := server.NewContext(app)
	if err != nil {
		return err
	}
	defer ctx.Close()

	names := app.Args()

	w, err := waiter.NewWaiter(ctx)
	if err != nil {
		return err
	}

	var lastErr error
	for _, name := range names {
		stack, err := lookup.Lookup(ctx.Client, name, client.StackType)
		if err != nil {
			return err
		}

		err = ctx.Client.Ops.DoDelete(stack.Links[clientbase.SELF])
		if err != nil {
			lastErr = err
			continue
		}

		w.Add(stack.ID)
	}

	if lastErr != nil {
		return lastErr
	}

	return w.Wait()
}

func stackUpdate(app *cli.Context) error {
	ctx, err := server.NewContext(app)
	if err != nil {
		return err
	}
	defer ctx.Close()

	if len(app.Args()) == 0 {
		return fmt.Errorf("at least on stack name is required")
	}

	names := app.Args()

	w, err := waiter.NewWaiter(ctx)
	if err != nil {
		return err
	}

	var lastErr error
	for _, name := range names {
		stack, err := lookup.Lookup(ctx.Client, name, client.StackType)
		if err != nil {
			return err
		}

		resp := &client.Stack{}
		err = ctx.Client.Ops.DoUpdate(client.StackType, stack, &client.Stack{
			Description: app.String("description"),
		}, resp)
		if err != nil {
			lastErr = err
			continue
		}

		w.Add(stack.ID)
	}

	if lastErr != nil {
		return lastErr
	}

	return w.Wait()
}

func defaultAction(fn func(ctx *cli.Context) error) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		if ctx.Bool("help") {
			cli.ShowAppHelp(ctx)
			return nil
		}
		return fn(ctx)
	}
}
