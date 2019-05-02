package ps

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/tables"
	"github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/serviceset"
	"k8s.io/apimachinery/pkg/runtime"
)

type ServiceData struct {
	ID       string
	Created  string
	Service  *riov1.Service
	Endpoint string
	External string
}

func (p *Ps) apps(ctx *clicontext.CLIContext) error {
	objs, err := ctx.List(types.AppType)
	if err != nil {
		return err
	}
	writer := tables.NewApp(ctx)
	return writer.Write(objs)
}

func (p *Ps) revisions(ctx *clicontext.CLIContext) error {
	objs, err := ctx.List(types.ServiceType)
	if err != nil {
		return err
	}

	var output []runtime.Object
	var services []*riov1.Service
	for _, obj := range objs {
		services = append(services, obj.(*riov1.Service))
	}
	set, err := serviceset.CollectionServices(services)
	if err != nil {
		return err
	}

	// list services for specific app
	if len(ctx.CLI.Args()) > 0 {
		for _, app := range ctx.CLI.Args() {
			revs, ok := set[app]
			if !ok {
				continue
			}
			for _, rev := range revs.Revisions {
				output = append(output, rev)
			}
		}
	} else {
		for _, rev := range set.List() {
			output = append(output, rev)
		}
	}

	writer := tables.NewService(ctx)
	return writer.Write(output)
}
