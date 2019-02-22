package externalservice

import (
	"fmt"
	"net"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/stack"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

type Create struct {
}

func (c *Create) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) < 2 {
		return errors.New("Incorrect usage. Example: `rio externalservice create NAME TARGET...`")
	}
	externalService := &riov1.ExternalService{}
	target := ctx.CLI.Args().Tail()[0]
	if strings.ContainsRune(target, ',') {
		ips := strings.Split(target, ",")
		for _, ip := range ips {
			externalService.Spec.IPAddresses = append(externalService.Spec.IPAddresses, ip)
		}
	} else if ip := net.ParseIP(target); ip != nil {
		externalService.Spec.IPAddresses = append(externalService.Spec.IPAddresses, target)
	} else if strings.ContainsRune(target, '.') {
		externalService.Spec.FQDN = target
	} else {
		externalService.Spec.Service = target
	}
	var err error
	externalService.Spec.ProjectName, externalService.Spec.StackName, externalService.Name, err = stack.ResolveSpaceStackForName(ctx, ctx.CLI.Args().Get(0))
	if err != nil {
		return err
	}
	client, err := ctx.KubeClient()
	if err != nil {
		return err
	}
	es, err := client.Rio.ExternalServices(externalService.Spec.StackName).Create(externalService)
	if err != nil {
		return err
	}
	fmt.Println(es.Name)
	return nil
}
