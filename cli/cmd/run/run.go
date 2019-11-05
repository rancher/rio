package run

import (
	"fmt"
	"os"
	"time"

	"github.com/onsi/ginkgo/reporters/stenographer/support/go-isatty"
	"github.com/rancher/rio/cli/cmd/attach"
	"github.com/rancher/rio/cli/cmd/create"
	"github.com/rancher/rio/cli/pkg/clicontext"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

type Run struct {
	create.Create
}

func (r *Run) Run(ctx *clicontext.CLIContext) error {
	service, err := r.RunCallback(ctx, func(service *riov1.Service) *riov1.Service {
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
		return attach.RunAttach(ctx, time.Minute, service.Name, "")
	}

	return nil
}
