package inspect

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/table"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	"github.com/urfave/cli"
)

var (
	InspectTypes = []string{
		clitypes.ServiceType,
		clitypes.ConfigType,
		clitypes.StackType,
		clitypes.RouterType,
		clitypes.VolumeType,
		clitypes.ExternalServiceType,
		clitypes.PodType,
		clitypes.FeatureType,
		clitypes.PublicDomainType,
	}
)

type Inspect struct {
	T_Type string `desc:"The specific type to inspect"`
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
	types := InspectTypes
	if i.T_Type != "" {
		types = []string{i.T_Type}
	}

	for _, arg := range ctx.CLI.Args() {
		r, err := lookup.Lookup(ctx, arg, types...)
		if err != nil {
			return err
		}

		t := r.Type
		cmd, err := ctx.KubectlCmd(r.Namespace, "get", t, r.Name, "-o", "yaml")
		if err != nil {
			return err
		}
		return cmd.Run()
	}

	return nil
}
