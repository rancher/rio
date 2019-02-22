package publicdomain

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/norman/pkg/kv"
	"github.com/rancher/rio/cli/pkg/clicontext"
	projectv1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
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

	client, err := ctx.KubeClient()
	if err != nil {
		return err
	}

	space, stack, targetName := resolveSpaceStackName(target, cluster.DefaultProjectName, cluster.DefaultStackName)
	domain := &projectv1.PublicDomain{
		Spec: projectv1.PublicDomainSpec{
			DomainName:        domainName,
			TargetProjectName: space,
			TargetStackName:   stack,
			TargetName:        targetName,
		},
	}
	domain, err = client.Project.PublicDomains("").Create(domain)
	if err != nil {
		return err
	}
	fmt.Println(domain.Name)
	return nil
}

func resolveSpaceStackName(name, defaultProject, defaultStack string) (string, string, string) {
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
	return defaultProject, stackName, name
}
