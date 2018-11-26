package backend

import (
	"context"

	"github.com/rancher/rio/controllers/backend/istio"
	"github.com/rancher/rio/controllers/backend/publicdomain"
	"github.com/rancher/rio/controllers/backend/secrets"
	"github.com/rancher/rio/controllers/backend/service"
	"github.com/rancher/rio/controllers/backend/stack"
	"github.com/rancher/rio/controllers/backend/stackdef"
	"github.com/rancher/rio/controllers/backend/volume"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	istio.Register(ctx, rContext)
	stackdef.Register(ctx, rContext)
	if err := stack.Register(ctx, rContext); err != nil {
		return err
	}
	if err := service.Register(ctx, rContext); err != nil {
		return err
	}
	if err := volume.Register(ctx, rContext); err != nil {
		return err
	}
	publicdomain.Register(ctx, rContext)
	secrets.Register(ctx, rContext)
	return nil
}
