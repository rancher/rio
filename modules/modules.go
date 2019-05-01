package modules

import (
	"context"

	"github.com/rancher/rio/modules/autoscale"
	"github.com/rancher/rio/modules/build"
	"github.com/rancher/rio/modules/istio"
	"github.com/rancher/rio/modules/monitoring"
	"github.com/rancher/rio/modules/service"
	"github.com/rancher/rio/modules/storage"
	"github.com/rancher/rio/modules/system"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rioContext *types.Context) error {
	if err := system.Register(ctx, rioContext); err != nil {
		return err
	}
	if err := storage.Register(ctx, rioContext); err != nil {
		return err
	}
	if err := service.Register(ctx, rioContext); err != nil {
		return err
	}
	if err := istio.Register(ctx, rioContext); err != nil {
		return err
	}
	if err := monitoring.Register(ctx, rioContext); err != nil {
		return err
	}
	if err := autoscale.Register(ctx, rioContext); err != nil {
		return err
	}
	if err := build.Register(ctx, rioContext); err != nil {
		return err
	}
	return nil
}
