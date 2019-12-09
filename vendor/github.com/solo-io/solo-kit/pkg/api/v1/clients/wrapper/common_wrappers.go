package wrapper

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

func NewClusterClient(base clients.ResourceClient, cluster string) *Client {
	return &Client{
		ResourceClient: base,
		ProcessResource: func(resource resources.Resource) {
			resources.UpdateMetadata(resource, func(meta *core.Metadata) {
				meta.Cluster = cluster
			})
		},
	}
}

// Convenience function for wrapping clients only if they point to remote clusters.
func NewClusterResourceClient(base clients.ResourceClient, cluster string) clients.ResourceClient {
	if cluster == "" {
		return base
	}
	return NewClusterClient(base, cluster)
}
