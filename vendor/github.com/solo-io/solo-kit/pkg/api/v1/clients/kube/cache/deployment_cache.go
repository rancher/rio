package cache

import (
	"context"
	"sync"
	"time"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/controller"

	kubeinformers "k8s.io/client-go/informers"
	kubelisters "k8s.io/client-go/listers/apps/v1"

	"k8s.io/client-go/kubernetes"
)

type KubeDeploymentCache interface {
	DeploymentLister() kubelisters.DeploymentLister
	Subscribe() <-chan struct{}
	Unsubscribe(<-chan struct{})
}

type kubeDeploymentCache struct {
	deploymentLister kubelisters.DeploymentLister

	cacheUpdatedWatchers      []chan struct{}
	cacheUpdatedWatchersMutex sync.Mutex
}

// This context should live as long as the cache is desired. i.e. if the cache is shared
// across clients, it should get a context that has a longer lifetime than the clients themselves
func NewKubeDeploymentCache(ctx context.Context, client kubernetes.Interface) (*kubeDeploymentCache, error) {
	resyncDuration := 12 * time.Hour
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(client, resyncDuration)

	deployments := kubeInformerFactory.Apps().V1().Deployments()

	k := &kubeDeploymentCache{
		deploymentLister: deployments.Lister(),
	}

	kubeController := controller.NewController("kube-plugin-controller",
		controller.NewLockingSyncHandler(k.updatedOccured),
		deployments.Informer(),
	)

	stop := ctx.Done()
	err := kubeController.Run(2, stop)
	if err != nil {
		return nil, err
	}

	return k, nil
}

func (k *kubeDeploymentCache) DeploymentLister() kubelisters.DeploymentLister {
	return k.deploymentLister
}

func (k *kubeDeploymentCache) Subscribe() <-chan struct{} {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	c := make(chan struct{}, 10)
	k.cacheUpdatedWatchers = append(k.cacheUpdatedWatchers, c)
	return c
}

func (k *kubeDeploymentCache) Unsubscribe(c <-chan struct{}) {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	for i, cacheUpdated := range k.cacheUpdatedWatchers {
		if cacheUpdated == c {
			k.cacheUpdatedWatchers = append(k.cacheUpdatedWatchers[:i], k.cacheUpdatedWatchers[i+1:]...)
			return
		}
	}
}

func (k *kubeDeploymentCache) updatedOccured() {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	for _, cacheUpdated := range k.cacheUpdatedWatchers {
		select {
		case cacheUpdated <- struct{}{}:
		default:
		}
	}
}
