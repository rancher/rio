package inspect

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	client2 "github.com/rancher/rio/types/client/space/v1beta1"
	"github.com/urfave/cli"
)

var (
	InspectTypes = []string{
		client.ServiceType,
		client.ConfigType,
		client.StackType,
		client.RouteSetType,
		client.VolumeType,
		client.ExternalServiceType,
		client2.PodType,
		client2.NodeType,
		client2.PublicDomainType,
	}
)

type Inspect struct {
	T_Type  string `desc:"The specific type to inspect"`
	L_Links bool   `desc:"Include links and actions in output"`
}

func (i *Inspect) Customize(cmd *cli.Command) {
	for _, f := range table.WriterFlags() {
		if f.GetName() == "format" {
			sf := f.(cli.StringFlag)
			sf.Value = "json"
			cmd.Flags = append(cmd.Flags, sf)
		}
	}
}

func (i *Inspect) Run(ctx *clicontext.CLIContext) error {
	workspaceClient, err := ctx.WorkspaceClient()
	if err != nil {
		return err
	}
	spaceClient, err := ctx.ClusterClient()
	if err != nil {
		return err
	}
	ctx.WC = workspaceClient
	ctx.SC = spaceClient
	for _, arg := range ctx.CLI.Args() {
		r, err := find(ctx, arg, i.T_Type, InspectTypes)
		if err != nil {
			return err
		}
		if r == nil {
			continue
		}

		if !i.L_Links {
			delete(r, "links")
			delete(r, "actions")
		}

		t := table.NewWriter(nil, ctx)
		t.Write(r)
		if err := t.Close(); err != nil {
			return err
		}
	}

	return nil
}

func find(c lookup.ClientLookup, arg, override string, types []string) (map[string]interface{}, error) {
	if len(override) > 0 {
		types = []string{override}
	}
	r, err := lookup.Lookup(c, arg, types...)
	if err == nil {
		client, err := c.ClientLookup(r.Type)
		if err != nil {
			return nil, err
		}
		data := map[string]interface{}{}
		err = client.GetLink(r.Resource, "self", &data)
		if err == nil {
			return data, nil
		}
	}
	return nil, nil
}
