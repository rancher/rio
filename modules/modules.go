package modules

import (
	"context"

	"github.com/rancher/rio/modules/istio"
	"github.com/rancher/rio/modules/service"
	"github.com/rancher/rio/modules/system"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rioContext *types.Context) error {
	if err := system.Register(ctx, rioContext); err != nil {
		return err
	}
	if err := istio.Register(ctx, rioContext); err != nil {
		return err
	}
	return service.Register(ctx, rioContext)
}
