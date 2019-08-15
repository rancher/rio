package stacks

import (
	"fmt"

	"github.com/rancher/mapper/convert"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/stack"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

type info struct {
}

func (i info) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) != 1 {
		return fmt.Errorf("exactly one argument is required")
	}

	ns, name := stack.NamespaceAndName(ctx, ctx.CLI.Args()[0])
	r, err := ctx.ByID(ns, name, types.StackType)
	if err != nil {
		return err
	}
	s := r.Object.(*riov1.Stack)

	m, err := convert.EncodeToMap(s.Spec)
	if err != nil {
		return err
	}

	output, err := table.FormatYAML(m)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}
