package tag

import (
	"github.com/docker/distribution/reference"
	"github.com/pkg/errors"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

type Tag struct {
}

func (t *Tag) Run(ctx *clicontext.CLIContext) error {
	args := ctx.CLI.Args()
	if len(args) != 2 {
		return errors.New("Incorrect usage. Example: `rio tag $STACK_NAME $REPO_TAG`")
	}
	wc, err := ctx.WorkspaceClient()
	if err != nil {
		return err
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
	_, err = reference.ParseNormalizedNamed(repoTag)
	if err != nil {
		return err
	}
	for _, t := range stack.RepoTags {
		if t == repoTag {
			return nil
		}
	}
	stack.RepoTags = append(stack.RepoTags, repoTag)
	_, err = wc.Stack.Replace(stack)
	return err
}
