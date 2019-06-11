package inspect

import (
	"errors"
	"strings"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/table"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	"github.com/urfave/cli"
)

var (
	InspectTypes = []string{
		clitypes.AppType,
		clitypes.ServiceType,
		clitypes.ConfigType,
		clitypes.NamespaceType,
		clitypes.RouterType,
		clitypes.ExternalServiceType,
		clitypes.PodType,
		clitypes.FeatureType,
		clitypes.PublicDomainType,
		clitypes.BuildType,
		clitypes.SecretType,
	}
)

type Inspect struct {
	T_Type string `desc:"The specific type to inspect"`
}

func (i *Inspect) Customize(cmd *cli.Command) {
	for _, f := range table.WriterFlags() {
		if f.GetName() == "format" {
			sf := f.(cli.StringFlag)
			sf.Value = "yaml"
			cmd.Flags = append(cmd.Flags, sf)
		}
	}
}

func (i *Inspect) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return errors.New("at least one argument is required")
	}

	types := InspectTypes
	if i.T_Type != "" {
		types = []string{i.T_Type}
	}

	for _, arg := range ctx.CLI.Args() {
		if strings.Contains(arg, ":") {
			types = []string{clitypes.ServiceType}
		} else {
			for i, t := range types {
				if t == clitypes.ServiceType {
					types = append(types[0:i], types[i+1:]...)
					break
				}
			}
		}
		r, err := lookup.Lookup(ctx, arg, types...)
		if err != nil {
			return err
		}

		t := table.NewWriter(nil, ctx)
		t.Write(r.Object)
		if err := t.Close(); err != nil {
			return err
		}
	}

	return nil
}
