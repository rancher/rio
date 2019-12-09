package cache

import (
	"context"
	"sync"
	"time"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/controller"
	"github.com/solo-io/solo-kit/pkg/multicluster/clustercache"
	"k8s.io/client-go/rest"

	kubeinformers "k8s.io/client-go/informers"
	kubelisters "k8s.io/client-go/listers/batch/v1"

	"k8s.io/client-go/kubernetes"
)

type KubeJobCache interface {
	clustercache.ClusterCache
	JobLister() kubelisters.JobLister
	Subscribe() <-chan struct{}
	Unsubscribe(<-chan struct{})
}

type kubeJobCache struct {
	jobLister kubelisters.JobLister

	cacheUpdatedWatchers      []chan struct{}
	cacheUpdatedWatchersMutex sync.Mutex
}

// This context should live as long as the cache is desired. i.e. if the cache is shared
// across clients, it should get a context that has a longer lifetime than the clients themselves
func NewKubeJobCache(ctx context.Context, client kubernetes.Interface) (*kubeJobCache, error) {
	resyncDuration := 12 * time.Hour
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(client, resyncDuration)

	jobs := kubeInformerFactory.Batch().V1().Jobs()

	k := &kubeJobCache{
		jobLister: jobs.Lister(),
	}

	kubeController := controller.NewController("kube-plugin-controller",
		controller.NewLockingSyncHandler(k.updatedOccured),
		jobs.Informer(),
	)

	stop := ctx.Done()
	err := kubeController.Run(2, stop)
	if err != nil {
		return nil, err
	}

	return k, nil
}

func NewJobCacheFromConfig(ctx context.Context, cluster string, restConfig *rest.Config) clustercache.ClusterCache {
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil
	}
	c, err := NewKubeJobCache(ctx, kubeClient)
	if err != nil {
		return nil
	}
	return c
}

var _ clustercache.NewClusterCacheForConfig = NewJobCacheFromConfig

func (k *kubeJobCache) JobLister() kubelisters.JobLister {
	return k.jobLister
}

func (k *kubeJobCache) Subscribe() <-chan struct{} {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	c := make(chan struct{}, 10)
	k.cacheUpdatedWatchers = append(k.cacheUpdatedWatchers, c)
	return c
}

func (k *kubeJobCache) Unsubscribe(c <-chan struct{}) {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	for i, cacheUpdated := range k.cacheUpdatedWatchers {
		if cacheUpdated == c {
			k.cacheUpdatedWatchers = append(k.cacheUpdatedWatchers[:i], k.cacheUpdatedWatchers[i+1:]...)
			return
		}
	}
}

func (k *kubeJobCache) IsClusterCache() {}

func (k *kubeJobCache) updatedOccured() {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	for _, cacheUpdated := range k.cacheUpdatedWatchers {
		select {
		case cacheUpdated <- struct{}{}:
		default:
		}
	}
}
