package controllers

import (
	"context"

	"github.com/rancher/rio/pkg/controllers/config"
	"github.com/rancher/rio/pkg/controllers/systemstatus"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	// Controllers
	if err := config.Register(ctx, rContext); err != nil {
		return err
	}
	return systemstatus.Register(ctx, rContext)
}
