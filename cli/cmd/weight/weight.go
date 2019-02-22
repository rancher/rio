package weight

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/norman/pkg/kv"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Weight struct {
}

func (w *Weight) Run(ctx *clicontext.CLIContext) error {
	project, err := ctx.Project()
	if err != nil {
		return err
	}

	client, err := ctx.KubeClient()
	if err != nil {
		return err
	}

	for _, arg := range ctx.CLI.Args() {
		name, scaleStr := kv.Split(arg, "=")
		scaleStr = strings.TrimSuffix(scaleStr, "%")

		if scaleStr == "" {
			return fmt.Errorf("weight params must be in the format of SERVICE=PERCENTAGE, for example: mystack/myservice=10%%")
		}
		scale, err := strconv.Atoi(scaleStr)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %v", arg, err)
		}

		resource, err := lookup.Lookup(ctx, name, types.ServiceType)
		if err != nil {
			return err
		}

		service, err := client.Rio.Services(project.Project.Name).Get(resource.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		service.Spec.Revision.Weight = scale

		if _, err := client.Rio.Services(project.Project.Name).Update(service); err != nil {
			return err
		}
	}

	return nil
}
