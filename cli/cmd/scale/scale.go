package scale

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/norman/pkg/kv"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Scale struct {
}

func (s *Scale) Run(ctx *clicontext.CLIContext) error {
	client, err := ctx.KubeClient()
	if err != nil {
		return err
	}

	for _, arg := range ctx.CLI.Args() {
		name, scaleStr := kv.Split(arg, "=")
		resource, err := lookup.Lookup(ctx, name, clitypes.ServiceType)
		if err != nil {
			return err
		}
		service, err := client.Rio.Services(resource.Namespace).Get(resource.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if strings.ContainsRune(scaleStr, '-') {
			min, max := kv.Split(scaleStr, "-")
			minScale, _ := strconv.Atoi(min)
			maxScale, _ := strconv.Atoi(max)
			concurrency := 10
			if service.Spec.AutoScale != nil {
				concurrency = int(service.Spec.AutoScale.Concurrency)
			}
			service.Spec.AutoScale.MinScale = minScale
			service.Spec.AutoScale.MaxScale = maxScale
			service.Spec.AutoScale.Concurrency = concurrency
		} else {
			scale, err := strconv.Atoi(scaleStr)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %v", arg, err)
			}
			service.Spec.Scale = scale
		}

		if _, err := client.Rio.Services(resource.Namespace).Update(service); err != nil {
			return err
		}
	}

	return nil
}
