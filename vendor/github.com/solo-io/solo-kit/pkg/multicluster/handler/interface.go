package handler

import "k8s.io/client-go/rest"

type ClusterHandler interface {
	ClusterAdded(cluster string, restConfig *rest.Config)
	ClusterRemoved(cluster string, restConfig *rest.Config)
}
