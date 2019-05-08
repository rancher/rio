package storage

import (
	"context"

	"github.com/rancher/rio/modules/storage/features/localstorage"
	"github.com/rancher/rio/modules/storage/features/nfs"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	if err := localstorage.Register(ctx, rContext); err != nil {
		return err
	}
	return nfs.Register(ctx, rContext)
}
