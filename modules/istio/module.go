package system

import (
	"context"

	"github.com/rancher/rio/modules/istio/features/routing"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	return routing.Register(ctx, rContext)
}
