package inspect

import (
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/cli/server"
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
		client2.PodType,
		client2.NodeType,
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

func (i *Inspect) Run(app *cli.Context) error {
	ctx, err := server.NewContext(app)
	if err != nil {
		return err
	}
	defer ctx.Close()

	for _, arg := range app.Args() {
		r := find(ctx.ClientLookup, arg, i.T_Type, InspectTypes)
		if r == nil {
			continue
		}

		if !i.L_Links {
			delete(r, "links")
			delete(r, "actions")
		}

		t := table.NewWriter(nil, app)
		t.Write(r)
		if err := t.Close(); err != nil {
			return err
		}
	}

	return nil
}

func find(c lookup.ClientLookup, arg, override string, types []string) map[string]interface{} {
	if len(override) > 0 {
		types = []string{override}
	}
	r, err := lookup.Lookup(c, arg, types...)
	if err == nil {
		data := map[string]interface{}{}
		err = c(r.Type).GetLink(*r, "self", &data)
		if err == nil {
			return data
		}
	}
	return nil
}
