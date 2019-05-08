package service

import (
	"context"

	"github.com/rancher/rio/modules/service/features"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	return features.Register(ctx, rContext)
}
