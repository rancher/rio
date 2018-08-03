package run

import (
	"time"

	"fmt"

	"github.com/rancher/rio/cli/cmd/attach"
	"github.com/rancher/rio/cli/cmd/create"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	"github.com/urfave/cli"
)

type Run struct {
	create.Create
	Scale int `desc:"scale" default:"1"`
}

func (r *Run) Run(app *cli.Context) error {
	service, err := r.RunCallback(app, func(service *client.Service) *client.Service {
		service.Scale = int64(r.Scale)
		return service
	})
	if err != nil {
		return err
	}

	if !r.Detach && service.OpenStdin && service.Tty {
		fmt.Println("Attaching...")
		return attach.RunAttach(app, time.Minute, true, true, service.ID)
	}

	return nil
}
