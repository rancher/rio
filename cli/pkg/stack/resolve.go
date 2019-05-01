package stack

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/rio/cli/pkg/clicontext"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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

	if stackName == "" {
		stackName = c.DefaultStackName
	}

	stack, err := c.Rio.Stacks(c.Namespace).Get(stackName, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return "", "", "", errors.Wrapf(err, "failed to determine stack")
	}

	var s *riov1.Stack
	if apierrors.IsNotFound(err) {
		s, err = c.Rio.Stacks(stackName).Create(&riov1.Stack{
			ObjectMeta: metav1.ObjectMeta{
				Name:      stackName,
				Namespace: c.Namespace,
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
