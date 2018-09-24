package api

import (
	"context"

	"github.com/rancher/rio/controllers/api/domain"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	domain.Register(ctx, rContext)
	return nil
}
