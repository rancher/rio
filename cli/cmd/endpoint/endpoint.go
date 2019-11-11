package endpoint

import (
	"fmt"

	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/cli/pkg/tables"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	services2 "github.com/rancher/rio/pkg/services"
	"github.com/urfave/cli"
)

func Endpoints(app *cli.App) cli.Command {
	ls := builder.Command(&Endpoint{},
		"List rio endpoints ",
		app.Name+" endpoint",
		"")
	return cli.Command{
		Name:      "endpoints",
		ShortName: "endpoint",
		Usage:     "Operations on endpoints",
		Category:  "SUB COMMANDS",
		Action:    clicontext.DefaultAction(ls.Action),
		Flags:     table.WriterFlags(),
	}
}

type Endpoint struct {
}

type Data struct {
	Namespace string
	Name      string
	Endpoints []string
}

func (e *Endpoint) Run(ctx *clicontext.CLIContext) error {
	services, err := ctx.List(clitypes.ServiceType)
	if err != nil {
		return err
	}

	routers, err := ctx.List(clitypes.RouterType)
	if err != nil {
		return err
	}

	var data []Data
	seen := map[string]bool{}
	for _, svc := range services {
		service := svc.(*riov1.Service)

		app, _ := services2.AppAndVersion(service)
		key := fmt.Sprintf("%s/%s", service.Namespace, app)
		if seen[key] {
			continue
		}
		endpoints := util.NormalizingEndpoints(service.Status.AppEndpoints, "")
		if len(endpoints) > 0 {
			seen[key] = true
			data = append(data, Data{
				Name:      app,
				Namespace: service.Namespace,
				Endpoints: endpoints,
			})
		}
	}

	for _, router := range routers {
		r := router.(*riov1.Router)

		endpoints := util.NormalizingEndpoints(r.Status.Endpoints, "")
		if len(endpoints) > 0 {
			data = append(data, Data{
				Name:      r.Name,
				Namespace: r.Namespace,
				Endpoints: endpoints,
			})
		}
	}

	writer := tables.NewEndpoint(ctx)
	defer writer.TableWriter().Close()
	for _, obj := range data {
		writer.TableWriter().Write(obj)
	}
	return writer.TableWriter().Err()
}
