package externalservice

import (
	"net"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/stack"
	"github.com/rancher/rio/cli/pkg/waiter"
	client "github.com/rancher/rio/types/client/rio/v1"
)

type Create struct {
}

func (c *Create) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) < 2 {
		return errors.New("Incorrect usage. Example: `rio externalservice create NAME TARGET...`")
	}
	externalService := &client.ExternalService{}
	target := ctx.CLI.Args().Tail()[0]
	if strings.ContainsRune(target, ',') {
		ips := strings.Split(target, ",")
		for _, ip := range ips {
			externalService.IPAddresses = append(externalService.IPAddresses, ip)
		}
	} else if ip := net.ParseIP(target); ip != nil {
		externalService.IPAddresses = append(externalService.IPAddresses, target)
	} else if strings.ContainsRune(target, '.') {
		externalService.FQDN = target
	} else {
		externalService.Service = target
	}
	var err error
	externalService.ProjectID, externalService.StackID, externalService.Name, err = stack.ResolveSpaceStackForName(ctx, ctx.CLI.Args().Get(0))
	if err != nil {
		return err
	}
	wc, err := ctx.ProjectClient()
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
