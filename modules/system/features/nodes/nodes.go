package nodes

import (
	"context"

	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/types"
	corev1 "k8s.io/api/core/v1"
)

const (
	indexName = "nodeEndpointIndexer"
)

func RegisterNodeEndpointIndexer(ctx context.Context, rContext *types.Context) error {
	i := indexer{
		namespace: rContext.Namespace,
	}
	rContext.Core.Core().V1().Endpoints().Cache().AddIndexer(indexName, i.indexEPByNode)
	return nil
}

type indexer struct {
	namespace string
}

func (i indexer) indexEPByNode(ep *corev1.Endpoints) ([]string, error) {
	if ep.Namespace != i.namespace || ep.Name != constants.GatewayName {
		return nil, nil
	}

	var result []string

	for _, subset := range ep.Subsets {
		for _, addr := range subset.Addresses {
			if addr.NodeName != nil {
				result = append(result, *addr.NodeName)
			}
		}
	}

	return result, nil
}
