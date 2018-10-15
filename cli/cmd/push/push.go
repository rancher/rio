package push

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/docker/app/pkg/resto"
	"github.com/pkg/errors"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/yamldownload"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

const rioStackName = "rio-stack.yaml"

type Push struct {
}

func (p *Push) Run(ctx *clicontext.CLIContext) error {
	wc, err := ctx.WorkspaceClient()
	if err != nil {
		return err
	}
	args := ctx.CLI.Args()

	if len(args) != 2 {
		return errors.New("Incorrect usage. Example: `rio push $STACK_NAME $REPO_TAG`")
	}

	stackName := args.Get(0)
	repoTag := args.Get(1)

	resource, err := lookup.Lookup(ctx, stackName, client.StackType)
	if err != nil {
		return err
	}
	stack, err := wc.Stack.ByID(resource.ID)
	if err != nil {
		return err
	}
	foundTag := false
	for _, t := range stack.RepoTags {
		if t == repoTag {
			foundTag = true
			break
		}
	}
	if !foundTag {
		return errors.Errorf("No tag %s found on stack %s. Try `rio tag` first", repoTag, stackName)
	}
	_, body, _, err := yamldownload.DownloadYAML(ctx, "yaml", "export", stackName, client.StackType)
	if err != nil {
		return err
	}
	defer body.Close()
	content, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	payload := map[string]string{
		rioStackName: string(content),
	}
	for fileName, additionFile := range stack.AdditionalFiles {
		payload[fileName] = additionFile
	}
	digest, err := resto.PushConfigMulti(context.Background(), payload, repoTag, resto.RegistryOptions{}, nil)
	if err != nil {
		return err
	}
	fmt.Println(digest)
	return nil
}
