package stacks

import (
	"fmt"

	"github.com/rancher/rio/cli/cmd/apply"
	"github.com/rancher/rio/cli/pkg/clicontext"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

type update struct {
	Answers string `desc:"Update answer file"`
}

func (u update) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) != 1 {
		return fmt.Errorf("exactly one argument is required")
	}

	r, err := ctx.ByID(ctx.CLI.Args()[0])
	if err != nil {
		return err
	}
	s := r.Object.(*riov1.Stack)

	answers, err := apply.ReadAnswers(u.Answers)
	if err != nil {
		return err
	}

	s.Spec.Answers = answers
	return ctx.UpdateObject(s)
}
