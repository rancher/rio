package externalservice

import (
	"github.com/pkg/errors"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/stack"
	"github.com/rancher/rio/cli/pkg/waiter"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

type Create struct {
}

func (c *Create) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) < 2 {
		return errors.New("Incorrect usage. Example: `rio externalservice add NAME TARGET...`")
	}
	externalService := &client.ExternalService{
		Target: ctx.CLI.Args().Tail()[0],
	}
	var err error
	externalService.SpaceID, externalService.StackID, externalService.Name, err = stack.ResolveSpaceStackForName(ctx, ctx.CLI.Args().Get(0))
	if err != nil {
		return err
	}
	wc, err := ctx.WorkspaceClient()
	if err != nil {
		return err
	}
	w, err := waiter.NewWaiter(ctx)
	if err != nil {
		return err
	}
	es, err := wc.ExternalService.Create(externalService)
	if err != nil {
		return err
	}
	w.Add(&es.Resource)
	return w.Wait(ctx.Ctx)
}
