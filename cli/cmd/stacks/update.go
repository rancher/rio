package stacks

import (
	"fmt"

	"github.com/rancher/rio/cli/cmd/apply"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/stack"
	"github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
)

type update struct {
	Answers string   `desc:"Update answer file"`
	Images  []string `desc:"Update image for services"`
}

func (u update) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) != 1 {
		return fmt.Errorf("exactly one argument is required")
	}

	ns, name := stack.NamespaceAndName(ctx, ctx.CLI.Args()[0])
	r, err := ctx.ByID(ns, name, types.StackType)
	if err != nil {
		return err
	}
	s := r.Object.(*riov1.Stack)

	answers, err := apply.ReadAnswers(u.Answers)
	if err != nil {
		return err
	}

	images := map[string]string{}
	for _, i := range u.Images {
		serviceName, image := kv.Split(i, "=")
		images[serviceName] = image
	}

	s.Spec.Answers = answers
	s.Spec.Images = images
	return ctx.UpdateObject(s)
}
