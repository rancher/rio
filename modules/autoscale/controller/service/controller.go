package service

import (
	"context"

	"github.com/rancher/rio/modules/autoscale/controller/service/populate"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "autoscale-service", rContext.Rio.Rio().V1().Service())
	c.Apply = c.Apply.WithCacheTypes(rContext.AutoScale.Autoscale().V1().ServiceScaleRecommendation())

	c.Populator = populate.ServiceRecommendationForService

	return nil
}
