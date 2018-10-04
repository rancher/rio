package run

import (
	"fmt"
	"os"
	"time"

	"github.com/onsi/ginkgo/reporters/stenographer/support/go-isatty"
	"github.com/rancher/rio/cli/cmd/attach"
	"github.com/rancher/rio/cli/cmd/create"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

type Run struct {
	create.Create
	Scale int `desc:"scale" default:"1"`
}

func (r *Run) Run(ctx *clicontext.CLIContext) error {
	service, err := r.RunCallback(ctx, func(service *client.Service) *client.Service {
		service.Scale = int64(r.Scale)
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
