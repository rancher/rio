package ps

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/tables"
	"github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
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
