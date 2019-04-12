package monitoring

import (
	"context"

	"github.com/rancher/rio/modules/monitoring/features/telemetry"

	"github.com/rancher/rio/modules/monitoring/features/prometheus"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	if err := prometheus.Register(ctx, rContext); err != nil {
		return err
	}

	if err := telemetry.Register(ctx, rContext); err != nil {
		return err
	}

	return nil
}
