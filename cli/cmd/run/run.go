package run

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rancher/norman/pkg/kv"

	isatty "github.com/onsi/ginkgo/reporters/stenographer/support/go-isatty"
	"github.com/rancher/rio/cli/cmd/attach"
	"github.com/rancher/rio/cli/cmd/create"
	"github.com/rancher/rio/cli/pkg/clicontext"
	client "github.com/rancher/rio/types/client/rio/v1"
)

type Run struct {
	create.Create
	Scale string `desc:"scale" default:"1"`
}

func (r *Run) Run(ctx *clicontext.CLIContext) error {
	service, err := r.RunCallback(ctx, func(service *client.Service) *client.Service {
		if strings.ContainsRune(r.Scale, '-') {
			min, max := kv.Split(r.Scale, "-")
			minScale, _ := strconv.Atoi(min)
			maxScale, _ := strconv.Atoi(max)
			service.AutoScale = &client.AutoscaleConfig{}
			service.AutoScale.MinScale = int64(minScale)
			service.AutoScale.MaxScale = int64(maxScale)
			service.AutoScale.Concurrency = int64(r.Concurrency)
			service.Scale = int64(minScale)
			return service
		}

		scale, _ := strconv.Atoi(r.Scale)
		if scale == 0 {
			scale = 1
		}
		service.Scale = int64(scale)
		return service
	})
	if err != nil {
		return err
	}

	istty := isatty.IsTerminal(os.Stdout.Fd()) &&
		isatty.IsTerminal(os.Stderr.Fd()) &&
		isatty.IsTerminal(os.Stdin.Fd())

	if istty && !r.Detach && service.OpenStdin && service.Tty {
		fmt.Println("Attaching...")
		return attach.RunAttach(ctx, time.Minute, true, true, service.ID)
	}

	return nil
}
