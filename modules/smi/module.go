package smi

import (
	"context"

	features "github.com/rancher/rio/modules/smi/feature"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	return features.Register(ctx, rContext)
}
