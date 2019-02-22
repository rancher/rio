package stack

import (
	"strings"

	errors2 "k8s.io/apimachinery/pkg/api/errors"

	"github.com/pkg/errors"
	"github.com/rancher/norman/pkg/kv"
	"github.com/rancher/rio/cli/pkg/clicontext"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	client, err := cluster.KubeClient()
	if err != nil {
		return "", "", "", err
	}

	p, err := c.Project()
	if err != nil {
		return "", "", "", err
	}

	if stackName == "" {
		stackName = cluster.DefaultStackName
	}

	stack, err := client.Rio.Stacks(p.Project.Name).Get(stackName, metav1.GetOptions{})
	if err != nil && !errors2.IsNotFound(err) {
		return "", "", "", errors.Wrapf(err, "failed to determine stack")
	}

	var s *riov1.Stack
	if errors2.IsNotFound(err) {
		s, err = client.Rio.Stacks(stackName).Create(&riov1.Stack{
			ObjectMeta: metav1.ObjectMeta{
				Name:      stackName,
				Namespace: p.Project.Name,
			},
		})
		if err != nil {
			return "", "", "", errors.Wrapf(err, "failed to create stack %s", stackName)
		}
	} else {
		s = stack
	}

	return s.Namespace, s.Name, name, nil
}
