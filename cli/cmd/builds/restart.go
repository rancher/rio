package builds

import (
	"fmt"
	"strings"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/stack"
	"github.com/rancher/wrangler/pkg/kv"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Restart struct {
}

func (r *Restart) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return fmt.Errorf("at least one argument is required")
	}

	serviceName, revision := kv.Split(ctx.CLI.Args()[0], ":")
	namespace, name := stack.NamespaceAndName(ctx, serviceName)

	builds, err := ctx.Build.Builds(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("service-name=%s, service-namespace=%s", name, namespace),
	})
	if err != nil {
		return err
	}

	for _, build := range builds.Items {
		if strings.Contains(build.Name, revision) {
			if err := ctx.Build.Builds(namespace).Delete(build.Name, &metav1.DeleteOptions{}); err != nil {
				return err
			}
		}
	}
	return nil
}
