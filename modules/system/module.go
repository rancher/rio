package system

import (
	"context"

	"github.com/rancher/rio/modules/system/features/rdns"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	return rdns.Register(ctx, rContext)
}
