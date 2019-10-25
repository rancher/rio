package controllers

import (
	"context"

	"github.com/rancher/rio/pkg/controllers/config"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	// Controllers
	return config.Register(ctx, rContext)
}
