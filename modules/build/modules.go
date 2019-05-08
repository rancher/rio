package build

import (
	"context"

	"github.com/rancher/rio/modules/build/feature"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	return feature.Register(ctx, rContext)
}
