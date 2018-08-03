package node

import (
	"github.com/rancher/norman/clientbase"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/cli/pkg/waiter"
	"github.com/rancher/rio/cli/server"
	spaceclient "github.com/rancher/rio/types/client/space/v1beta1"
	"github.com/urfave/cli"
)

func Node() cli.Command {
	return cli.Command{
		Name:      "nodes",
		ShortName: "node",
		Usage:     "Operations on nodes",
		Action:    defaultAction(nodeLs),
		Flags:     table.WriterFlags(),
		Category:  "SUB COMMANDS",
		Subcommands: []cli.Command{
			{
				Name:      "ls",
				Usage:     "List nodes",
				ArgsUsage: "None",
				Action:    nodeLs,
				Flags:     table.WriterFlags(),
			},
			{
				Name:      "delete",
				ShortName: "rm",
				Usage:     "Delete a node",
				ArgsUsage: "None",
				Action:    nodeRm,
			},
		},
	}
}

type Data struct {
	ID   string
	Node spaceclient.Node
}

func nodeLs(app *cli.Context) error {
	ctx, err := server.NewContext(app)
	if err != nil {
		return err
	}
	defer ctx.Close()

	c, err := ctx.SpaceClient()
	if err != nil {
		return err
	}

	collection, err := c.Node.List(nil)
	if err != nil {
		return err
	}

	writer := table.NewWriter([][]string{
		{"NAME", "{{.Node | name}}"},
		{"STATE", "Node.State"},
		{"ADDRESS", "{{.Node | address}}"},
		{"DETAIL", "Node.TransitioningMessage"},
	}, app)
	defer writer.Close()

	writer.AddFormatFunc("address", FormatAddress)
	writer.AddFormatFunc("name", FormatName)

	for _, item := range collection.Data {
		writer.Write(&Data{
			ID:   item.ID,
			Node: item,
		})
	}

	return writer.Err()
}

func nodeRm(app *cli.Context) error {
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
		node, err := lookup.Lookup(ctx.Client, name, spaceclient.NodeType)
		if err != nil {
			return err
		}

		err = ctx.Client.Ops.DoDelete(node.Links[clientbase.SELF])
		if err != nil {
			lastErr = err
			continue
		}

		w.Add(node.ID)
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

func FormatAddress(data interface{}) (string, error) {
	node, ok := data.(spaceclient.Node)
	if !ok {
		return "", nil
	}

	internalIP := ""
	externalIP := ""

	for _, addr := range node.Addresses {
		if addr.Type == "InternalIP" {
			internalIP = addr.Address
		} else if addr.Type == "ExternalIP" {
			externalIP = addr.Address
		}
	}

	addr := internalIP
	if externalIP != "" {
		if addr == "" {
			addr = externalIP
		}
		addr += "/" + externalIP
	}

	return addr, nil
}

func FormatName(data interface{}) (string, error) {
	node, ok := data.(spaceclient.Node)
	if !ok {
		return "", nil
	}

	name := node.Name
	for _, addr := range node.Addresses {
		if addr.Type == "Hostname" {
			name = addr.Address
		}
	}

	return name, nil
}
