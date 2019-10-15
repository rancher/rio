package stats

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

func IncrementResourceCount(ctx context.Context, namespace, resource string, m *stats.Int64Measure) {
	if err := stats.RecordWithTags(
		ctx,
		[]tag.Mutator{
			tag.Insert(NamespaceKey, namespace),
			tag.Insert(ResourceKey, resource),
		},
		m.M(1),
	); err != nil {
		contextutils.LoggerFrom(ctx).Errorf("incrementing resource count: %v", err)
	}
}
