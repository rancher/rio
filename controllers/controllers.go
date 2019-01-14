package controllers

import (
	"context"

	"github.com/rancher/rio/features/kiali"

	"github.com/rancher/rio/features/prometheus"

	"github.com/rancher/rio/controllers/data"
	"github.com/rancher/rio/controllers/feature"
	"github.com/rancher/rio/controllers/setting"
	"github.com/rancher/rio/features/autoscaling"
	"github.com/rancher/rio/features/letsencrypt"
	"github.com/rancher/rio/features/localstorage"
	"github.com/rancher/rio/features/monitoring"
	"github.com/rancher/rio/features/nfs"
	"github.com/rancher/rio/features/rdns"
	"github.com/rancher/rio/features/routing"
	"github.com/rancher/rio/features/stack"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	// Features
	if err := stack.Register(ctx, rContext); err != nil {
		return err
	}
	if err := letsencrypt.Register(ctx, rContext); err != nil {
		return err
	}
	if err := nfs.Register(ctx, rContext); err != nil {
		return err
	}
	if err := monitoring.Register(ctx, rContext); err != nil {
		return err
	}
	if err := routing.Register(ctx, rContext); err != nil {
		return err
	}
	if err := rdns.Register(ctx, rContext); err != nil {
		return err
	}
	if err := localstorage.Register(ctx, rContext); err != nil {
		return err
	}
	if err := autoscaling.Register(ctx, rContext); err != nil {
		return err
	}
	if err := prometheus.Register(ctx, rContext); err != nil {
		return err
	}
	if err := kiali.Register(ctx, rContext); err != nil {
		return err
	}

	// Controllers
	if err := data.Register(ctx, rContext); err != nil {
		return err
	}
	if err := feature.Register(ctx, rContext); err != nil {
		return err
	}
	if err := setting.Register(ctx, rContext); err != nil {
		return err
	}

	return nil
}
