package stack

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/norman/pkg/kv"
	"github.com/rancher/norman/types"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/waiter"
	"github.com/rancher/rio/types/client/rio/v1"
)

func ResolveSpaceStackForName(c *clicontext.CLIContext, in string) (string, string, string, error) {
	stackName, name := kv.Split(in, "/")
	if stackName != "" && name == "" {
		if !strings.HasSuffix(in, "/") {
			name = stackName
			stackName = ""
		}
	}

	cluster, err := c.Cluster()
	if err != nil {
		return "", "", "", err
	}

	w, err := c.Project()
	if err != nil {
		return "", "", "", err
	}

	wc, err := w.Client()
	if err != nil {
		return "", "", "", err
	}

	if stackName == "" {
		stackName = cluster.DefaultStackName
	}

	stacks, err := wc.Stack.List(&types.ListOpts{
		Filters: map[string]interface{}{
			"name": stackName,
		},
	})
	if err != nil {
		return "", "", "", errors.Wrapf(err, "failed to determine stack")
	}

	var s *client.Stack
	if len(stacks.Data) == 0 {
		s, err = wc.Stack.Create(&client.Stack{
			Name:      stackName,
			ProjectID: w.ID,
		})
		if err != nil {
			return "", "", "", errors.Wrapf(err, "failed to create stack %s", stackName)
		}
	} else {
		s = &stacks.Data[0]
	}

	if err := waiter.EnsureActive(c, &s.Resource); err != nil {
		return "", "", "", err
	}

	return s.ProjectID, s.ID, name, nil
}
