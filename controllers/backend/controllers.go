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
	stack.Register(ctx, rContext)
	service.Register(ctx, rContext)
	volume.Register(ctx, rContext)
	publicdomain.Register(ctx, rContext)
	secrets.Register(ctx, rContext)
	return nil
}
