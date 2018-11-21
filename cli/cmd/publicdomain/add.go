package publicdomain

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/norman/pkg/kv"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/types/client/space/v1beta1"
)

type Add struct {
}

func (a *Add) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) != 2 {
		return errors.New("Incorrect Usage. Example: `rio domain add DOMAIN_NAME TARGET_SVC`")
	}
	domainName := ctx.CLI.Args().Get(0)
	target := ctx.CLI.Args().Get(1)
	cluster, err := ctx.Cluster()
	if err != nil {
		return err
	}
	spaceClient, err := cluster.Client()
	if err != nil {
		return err
	}
	space, stack, targetName := resolveSpaceStackName(target, cluster.DefaultWorkspaceName, cluster.DefaultStackName)
	domain := &client.PublicDomain{
		DomainName:          domainName,
		TargetWorkspaceName: space,
		TargetStackName:     stack,
		TargetName:          targetName,
	}
	domain, err = spaceClient.PublicDomain.Create(domain)
	if err != nil {
		return err
	}
	fmt.Println(domain.Name)
	return nil
}

func resolveSpaceStackName(name, defaultWorkspace, defaultStack string) (string, string, string) {
	parts := strings.SplitN(name, "/", 3)
	if len(parts) == 3 {
		return parts[0], parts[1], parts[2]
	}
	stackName, name := kv.Split(name, "/")
	if stackName != "" && name == "" {
		if !strings.HasSuffix(name, "/") {
			name = stackName
			stackName = ""
		}
	}
	if stackName == "" {
		stackName = defaultStack
	}
	return defaultWorkspace, stackName, name
}
