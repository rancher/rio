package endpoint

import (
	"net/url"
	"sort"

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
		if seen[app] {
			continue
		}
		var endpoints []string
		hostNameSeen := map[string]string{}
		for _, e := range service.Status.AppEndpoints {
			u, _ := url.Parse(e)
			if u.Scheme == "https" {
				hostNameSeen[u.Hostname()] = e
			} else {
				if _, ok := hostNameSeen[u.Hostname()]; !ok {
					hostNameSeen[u.Hostname()] = e
				}
			}
		}

		for _, v := range hostNameSeen {
			endpoints = append(endpoints, v)
		}
		sort.Strings(endpoints)

		if len(endpoints) > 0 {
			data = append(data, Data{
				Name:      app,
				Endpoints: endpoints,
			})
		}
	}

	for _, router := range routers {
		r := router.(*riov1.Router)

		var endpoints []string
		hostNameSeen := map[string]string{}
		for _, e := range r.Status.Endpoints {
			u, _ := url.Parse(e)
			if u.Scheme == "https" {
				hostNameSeen[u.Hostname()] = e
			} else {
				if _, ok := hostNameSeen[u.Hostname()]; !ok {
					hostNameSeen[u.Hostname()] = e
				}
			}
		}

		for _, v := range hostNameSeen {
			endpoints = append(endpoints, v)
		}
		sort.Strings(endpoints)

		if len(endpoints) > 0 {
			data = append(data, Data{
				Name:      r.Name,
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
