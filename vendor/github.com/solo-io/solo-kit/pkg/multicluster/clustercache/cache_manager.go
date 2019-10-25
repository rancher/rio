package clustercache

import (
	"context"
	"sync"

	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/multicluster/handler"
	"k8s.io/client-go/rest"
)

//go:generate mockgen -destination=./mocks/cache_manager.go -source cache_manager.go -package mocks

type cacheWrapper struct {
	cancel    context.CancelFunc
	coreCache ClusterCache
}

// PerClusterCaches are caches that can be created from a *rest.Config and shared per-cluster
// by the CacheManager. All kube caches should be PerClusterCaches so that we can maintain
// exactly one cache per registered cluster.
type ClusterCache interface {
	IsClusterCache()
}

type NewClusterCacheForConfig func(ctx context.Context, cluster string, restConfig *rest.Config) ClusterCache

type CacheGetter interface {
	GetCache(cluster string, restConfig *rest.Config) ClusterCache
}

type CacheManager interface {
	handler.ClusterHandler
	CacheGetter
}

type manager struct {
	ctx          context.Context
	caches       map[string]cacheWrapper
	cacheAccess  sync.RWMutex
	newForConfig NewClusterCacheForConfig
}

var _ CacheManager = &manager{}

func NewCacheManager(ctx context.Context, newForConfig NewClusterCacheForConfig) (*manager, error) {
	if newForConfig == nil {
		return nil, errors.Errorf("cache manager requires a callback for generating per-cluster caches")
	}

	return &manager{
		ctx:          ctx,
		caches:       make(map[string]cacheWrapper),
		cacheAccess:  sync.RWMutex{},
		newForConfig: newForConfig,
	}, nil
}

func (m *manager) ClusterAdded(cluster string, restConfig *rest.Config) {
	// noop -- new caches are added lazily via GetCache
}

func (m *manager) addCluster(cluster string, restConfig *rest.Config) cacheWrapper {
	ctx, cancel := context.WithCancel(m.ctx)
	cw := cacheWrapper{
		cancel:    cancel,
		coreCache: m.newForConfig(ctx, cluster, restConfig),
	}
	m.cacheAccess.Lock()
	defer m.cacheAccess.Unlock()
	m.caches[cluster] = cw
	return cw
}

func (m *manager) ClusterRemoved(cluster string, restConfig *rest.Config) {
	m.cacheAccess.Lock()
	defer m.cacheAccess.Unlock()
	if cacheWrapper, exists := m.caches[cluster]; exists {
		cacheWrapper.cancel()
		delete(m.caches, cluster)
	}
}

func (m *manager) GetCache(cluster string, restConfig *rest.Config) ClusterCache {
	m.cacheAccess.RLock()
	cw, exists := m.caches[cluster]
	m.cacheAccess.RUnlock()
	if exists {
		return cw.coreCache
	}
	return m.addCluster(cluster, restConfig).coreCache
}
