package monitoring

import (
	"context"

	"github.com/rancher/rio/modules/monitoring/features/grafana"
	"github.com/rancher/rio/modules/monitoring/features/kiali"
	"github.com/rancher/rio/modules/monitoring/features/telemetry"

	"github.com/rancher/rio/modules/monitoring/features/prometheus"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	if err := prometheus.Register(ctx, rContext); err != nil {
		return err
	}

	if err := grafana.Register(ctx, rContext); err != nil {
		return err
	}

	if err := kiali.Register(ctx, rContext); err != nil {
		return err
	}

	return telemetry.Register(ctx, rContext)
}
