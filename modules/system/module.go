package system

import (
	"context"

	"github.com/rancher/rio/modules/system/features/letsencrypt"

	"github.com/rancher/rio/modules/system/features/rdns"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	if err := rdns.Register(ctx, rContext); err != nil {
		return err
	}
	return letsencrypt.Register(ctx, rContext)
}
