package backend

import (
	"context"

	"github.com/rancher/rio/controllers/backend/gateway"
	"github.com/rancher/rio/controllers/backend/pod"
	"github.com/rancher/rio/controllers/backend/service"
	"github.com/rancher/rio/controllers/backend/stack"
	"github.com/rancher/rio/controllers/backend/stackdeploy"
	"github.com/rancher/rio/controllers/backend/volume"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	gateway.Register(ctx, rContext)
	stack.Register(ctx, rContext)
	stackdeploy.Register(ctx, rContext)
	service.Register(ctx, rContext)
	pod.Register(ctx, rContext)
	volume.Register(ctx, rContext)
	return nil
}
