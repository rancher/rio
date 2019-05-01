package controllers

import (
	"context"

	"github.com/rancher/rio/pkg/controllers/feature"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	// Controllers
	if err := feature.Register(ctx, rContext); err != nil {
		return err
	}
	return nil
}
