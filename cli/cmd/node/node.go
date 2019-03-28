package node

import (
	"os"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/mapper"
	"github.com/rancher/rio/cli/pkg/table"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	"github.com/urfave/cli"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Node() cli.Command {
	return cli.Command{
		Name:      "nodes",
		ShortName: "node",
		Usage:     "Operations on nodes",
		Action:    clicontext.DefaultAction(nodeLs),
		Flags:     table.WriterFlags(),
		Category:  "SUB COMMANDS",
		Subcommands: []cli.Command{
			{
				Name:      "ls",
				Usage:     "List nodes",
				ArgsUsage: "None",
				Action:    clicontext.Wrap(nodeLs),
				Flags:     table.WriterFlags(),
			},
			{
				Name:      "delete",
				ShortName: "rm",
				Usage:     "Delete a node",
				ArgsUsage: "None",
				Action:    clicontext.Wrap(nodeRm),
			},
		},
	}
}

type Data struct {
	ID   string
	Node v1.Node
}

func nodeLs(ctx *clicontext.CLIContext) error {
	client, err := ctx.KubeClient()
	if err != nil {
		return err
	}

	collection, err := client.Core.Nodes("").List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	writer := table.NewWriter([][]string{
		{"NAME", "{{.Node | name}}"},
		{"STATE", "{{.Node | toJson | state}}"},
		{"ADDRESS", "{{.Node | address}}"},
		{"DETAIL", "{{.Node | toJson | transitioning}}"},
	}, ctx, os.Stdout)
	defer writer.Close()

	wrapper := mapper.GenericStatusMapper
	writer.AddFormatFunc("address", FormatAddress)
	writer.AddFormatFunc("name", FormatName)
	writer.AddFormatFunc("state", wrapper.FormatState)
	writer.AddFormatFunc("transitioning", wrapper.FormatTransitionMessage)

	for _, item := range collection.Items {
		writer.Write(&Data{
			ID:   item.Name,
			Node: item,
		})
	}

	return writer.Err()
}

func nodeRm(ctx *clicontext.CLIContext) error {
	names := ctx.CLI.Args()

	client, err := ctx.KubeClient()
	if err != nil {
		return err
	}

	var lastErr error
	for _, name := range names {
		node, err := lookup.Lookup(ctx, name, clitypes.NodeType)
		if err != nil {
			return err
		}

		if err := client.Core.Nodes("").Delete(node.Name, &metav1.DeleteOptions{}); err != nil {
			lastErr = err
		}
	}

	if lastErr != nil {
		return lastErr
	}
	return nil
}

func FormatAddress(data interface{}) (string, error) {
	node, ok := data.(v1.Node)
	if !ok {
		return "", nil
	}

	internalIP := ""
	externalIP := ""

	for _, addr := range node.Status.Addresses {
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
	node, ok := data.(v1.Node)
	if !ok {
		return "", nil
	}

	name := node.Name
	for _, addr := range node.Status.Addresses {
		if addr.Type == "Hostname" {
			name = addr.Address
		}
	}

	return name, nil
}
