package run

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	isatty "github.com/onsi/ginkgo/reporters/stenographer/support/go-isatty"
	"github.com/rancher/rio/cli/cmd/attach"
	"github.com/rancher/rio/cli/cmd/create"
	"github.com/rancher/rio/cli/pkg/clicontext"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
)

type Run struct {
	create.Create
	Scale string `desc:"scale" default:"1-10"`
}

func (r *Run) Run(ctx *clicontext.CLIContext) error {
	service, err := r.RunCallback(ctx, func(service *riov1.Service) *riov1.Service {
		if strings.ContainsRune(r.Scale, '-') {
			min, max := kv.Split(r.Scale, "-")
			if min != "" && max != "" {
				minScale, _ := strconv.Atoi(min)
				maxScale, _ := strconv.Atoi(max)
				service.Spec.Autoscale.MinReplicas = &minScale
				service.Spec.Autoscale.MaxReplicas = &maxScale
				service.Spec.Autoscale.Concurrency = &r.Concurrency
				if minScale != 0 {
					service.Spec.Replicas = &minScale
				} else {
					service.Spec.Replicas = &[]int{1}[0]
				}

				return service
			}
		}

		// disable autoscaling
		scale, _ := strconv.Atoi(r.Scale)
		service.Spec.Replicas = &scale
		service.Spec.Autoscale.MinReplicas = &scale
		service.Spec.Autoscale.MaxReplicas = &scale
		return service
	})
	if err != nil {
		return err
	}

	istty := isatty.IsTerminal(os.Stdout.Fd()) &&
		isatty.IsTerminal(os.Stderr.Fd()) &&
		isatty.IsTerminal(os.Stdin.Fd())

	if istty && service.Spec.Stdin && service.Spec.TTY {
		fmt.Println("Attaching...")
		return attach.RunAttach(ctx, time.Minute, service.Name)
	}

	return nil
}
