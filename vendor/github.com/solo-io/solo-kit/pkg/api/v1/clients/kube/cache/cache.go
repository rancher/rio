package cache

import (
	"context"
	"sync"
	"time"

	"github.com/solo-io/go-utils/stringutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/controller"
	"github.com/solo-io/solo-kit/pkg/errors"
	"go.opencensus.io/tag"

	v1 "k8s.io/api/core/v1"
	kubelisters "k8s.io/client-go/listers/core/v1"

	skkube "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type ServiceLister interface {
	// List lists all Services in the indexer.
	List(selector labels.Selector) (ret []*v1.Service, err error)
}

type PodLister interface {
	// List lists all Pods in the indexer.
	List(selector labels.Selector) (ret []*v1.Pod, err error)
}

type ConfigMapLister interface {
	// List lists all ConfigMaps in the indexer.
	List(selector labels.Selector) (ret []*v1.ConfigMap, err error)
}

type SecretLister interface {
	// List lists all Secrets in the indexer.
	List(selector labels.Selector) (ret []*v1.Secret, err error)
}

type Cache interface {
	Subscribe() <-chan struct{}
	Unsubscribe(<-chan struct{})
}

type KubeCoreCache interface {
	Cache

	// Deprecated: Use NamespacedPodLister instead
	PodLister() kubelisters.PodLister
	// Deprecated: Use NamespacedServiceLister instead
	ServiceLister() kubelisters.ServiceLister
	// Deprecated: Use NamespacedConfigMapLister instead
	ConfigMapLister() kubelisters.ConfigMapLister
	// Deprecated: Use NamespacedSecretLister instead
	SecretLister() kubelisters.SecretLister

	NamespaceLister() kubelisters.NamespaceLister

	NamespacedPodLister(ns string) PodLister
	NamespacedServiceLister(ns string) ServiceLister
	NamespacedConfigMapLister(ns string) ConfigMapLister
	NamespacedSecretLister(ns string) SecretLister
}

type kubeCoreCaches struct {
	podListers       map[string]kubelisters.PodLister
	serviceListers   map[string]kubelisters.ServiceLister
	configMapListers map[string]kubelisters.ConfigMapLister
	secretListers    map[string]kubelisters.SecretLister
	namespaceLister  kubelisters.NamespaceLister

	cacheUpdatedWatchers      []chan struct{}
	cacheUpdatedWatchersMutex sync.Mutex
}

// This context should live as long as the cache is desired. i.e. if the cache is shared
// across clients, it should get a context that has a longer lifetime than the clients themselves
func NewKubeCoreCache(ctx context.Context, client kubernetes.Interface) (*kubeCoreCaches, error) {
	resyncDuration := 12 * time.Hour
	return NewKubeCoreCacheWithOptions(ctx, client, resyncDuration, []string{metav1.NamespaceAll})
}

func NewKubeCoreCacheWithOptions(ctx context.Context, client kubernetes.Interface, resyncDuration time.Duration, namesapcesToWatch []string) (*kubeCoreCaches, error) {

	if len(namesapcesToWatch) == 0 {
		namesapcesToWatch = []string{metav1.NamespaceAll}
	}

	if len(namesapcesToWatch) > 1 {
		if stringutils.ContainsString(metav1.NamespaceAll, namesapcesToWatch) {
			return nil, errors.Errorf("if metav1.NamespaceAll is provided, it must be the only one. namespaces provided %v", namesapcesToWatch)
		}
	}

	var informers []cache.SharedIndexInformer

	pods := map[string]kubelisters.PodLister{}
	services := map[string]kubelisters.ServiceLister{}
	configMaps := map[string]kubelisters.ConfigMapLister{}
	secrets := map[string]kubelisters.SecretLister{}

	for _, nsToWatch := range namesapcesToWatch {
		nsToWatch := nsToWatch
		nsCtx := ctx
		if ctxWithTags, err := tag.New(nsCtx, tag.Insert(skkube.KeyNamespaceKind, skkube.NotEmptyValue(nsToWatch))); err == nil {
			nsCtx = ctxWithTags
		}

		{
			var typeCtx = nsCtx
			if ctxWithTags, err := tag.New(nsCtx, tag.Insert(skkube.KeyKind, "Pods")); err == nil {
				typeCtx = ctxWithTags
			}
			// Pods
			watch := client.CoreV1().Pods(nsToWatch).Watch
			list := func(options metav1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().Pods(nsToWatch).List(options)
			}
			informer := skkube.NewSharedInformer(typeCtx, resyncDuration, &v1.Pod{}, list, watch)
			informers = append(informers, informer)
			lister := kubelisters.NewPodLister(informer.GetIndexer())
			pods[nsToWatch] = lister
		}
		{
			var typeCtx = nsCtx
			if ctxWithTags, err := tag.New(nsCtx, tag.Insert(skkube.KeyKind, "Services")); err == nil {
				typeCtx = ctxWithTags
			}
			// Services
			watch := client.CoreV1().Services(nsToWatch).Watch
			list := func(options metav1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().Services(nsToWatch).List(options)
			}
			informer := skkube.NewSharedInformer(typeCtx, resyncDuration, &v1.Service{}, list, watch)
			informers = append(informers, informer)
			lister := kubelisters.NewServiceLister(informer.GetIndexer())
			services[nsToWatch] = lister
		}
		{
			var typeCtx = nsCtx
			if ctxWithTags, err := tag.New(nsCtx, tag.Insert(skkube.KeyKind, "ConfigMap")); err == nil {
				typeCtx = ctxWithTags
			}
			// ConfigMap
			watch := client.CoreV1().ConfigMaps(nsToWatch).Watch
			list := func(options metav1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().ConfigMaps(nsToWatch).List(options)
			}
			informer := skkube.NewSharedInformer(typeCtx, resyncDuration, &v1.ConfigMap{}, list, watch)
			informers = append(informers, informer)
			lister := kubelisters.NewConfigMapLister(informer.GetIndexer())
			configMaps[nsToWatch] = lister
		}
		{
			var typeCtx = nsCtx
			if ctxWithTags, err := tag.New(nsCtx, tag.Insert(skkube.KeyKind, "Secrets")); err == nil {
				typeCtx = ctxWithTags
			}
			// Secrets
			watch := client.CoreV1().Secrets(nsToWatch).Watch
			list := func(options metav1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().Secrets(nsToWatch).List(options)
			}
			informer := skkube.NewSharedInformer(typeCtx, resyncDuration, &v1.Secret{}, list, watch)
			informers = append(informers, informer)
			lister := kubelisters.NewSecretLister(informer.GetIndexer())
			secrets[nsToWatch] = lister
		}

	}

	var namespaceLister kubelisters.NamespaceLister
	if len(namesapcesToWatch) == 1 && namesapcesToWatch[0] == metav1.NamespaceAll {

		// Pods
		watch := client.CoreV1().Namespaces().Watch
		list := func(options metav1.ListOptions) (runtime.Object, error) {
			return client.CoreV1().Namespaces().List(options)
		}
		nsCtx := ctx
		if ctxWithTags, err := tag.New(nsCtx, tag.Insert(skkube.KeyNamespaceKind, skkube.NotEmptyValue(metav1.NamespaceAll)), tag.Insert(skkube.KeyKind, "Namespaces")); err == nil {
			nsCtx = ctxWithTags
		}
		informer := skkube.NewSharedInformer(nsCtx, resyncDuration, &v1.Namespace{}, list, watch)
		informers = append(informers, informer)
		namespaceLister = kubelisters.NewNamespaceLister(informer.GetIndexer())
	}

	k := &kubeCoreCaches{
		podListers:       pods,
		serviceListers:   services,
		configMapListers: configMaps,
		secretListers:    secrets,
		namespaceLister:  namespaceLister,
	}

	kubeController := controller.NewController("kube-plugin-controller",
		controller.NewLockingSyncHandler(k.updatedOccured), informers...,
	)

	stop := ctx.Done()
	err := kubeController.Run(2, stop)
	if err != nil {
		return nil, err
	}

	return k, nil
}

// Deprecated: Use NamespacedPodLister instead
func (k *kubeCoreCaches) PodLister() kubelisters.PodLister {
	return k.podListers[metav1.NamespaceAll]
}

// Deprecated: Use NamespacedServiceLister instead
func (k *kubeCoreCaches) ServiceLister() kubelisters.ServiceLister {
	return k.serviceListers[metav1.NamespaceAll]
}

// Deprecated: Use NamespacedConfigMapLister instead
func (k *kubeCoreCaches) ConfigMapLister() kubelisters.ConfigMapLister {
	return k.configMapListers[metav1.NamespaceAll]
}

// Deprecated: Use NamespacedSecretLister instead
func (k *kubeCoreCaches) SecretLister() kubelisters.SecretLister {
	return k.secretListers[metav1.NamespaceAll]
}

// NamespaceLister() will return a non-null lister only if we watch all namespaces.
func (k *kubeCoreCaches) NamespaceLister() kubelisters.NamespaceLister {
	return k.namespaceLister
}

func (k *kubeCoreCaches) NamespacedPodLister(ns string) PodLister {
	if lister, ok := k.podListers[metav1.NamespaceAll]; ok {
		return lister.Pods(ns)
	}
	return k.podListers[ns]
}

func (k *kubeCoreCaches) NamespacedServiceLister(ns string) ServiceLister {
	if lister, ok := k.serviceListers[metav1.NamespaceAll]; ok {
		return lister.Services(ns)
	}
	return k.serviceListers[ns]
}

func (k *kubeCoreCaches) NamespacedConfigMapLister(ns string) ConfigMapLister {
	if lister, ok := k.configMapListers[metav1.NamespaceAll]; ok {
		return lister.ConfigMaps(ns)
	}
	return k.configMapListers[ns]
}

func (k *kubeCoreCaches) NamespacedSecretLister(ns string) SecretLister {
	if lister, ok := k.secretListers[metav1.NamespaceAll]; ok {
		return lister.Secrets(ns)
	}
	return k.secretListers[ns]
}

func (k *kubeCoreCaches) Subscribe() <-chan struct{} {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	c := make(chan struct{}, 10)
	k.cacheUpdatedWatchers = append(k.cacheUpdatedWatchers, c)
	return c
}

func (k *kubeCoreCaches) Unsubscribe(c <-chan struct{}) {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	for i, cacheUpdated := range k.cacheUpdatedWatchers {
		if cacheUpdated == c {
			k.cacheUpdatedWatchers = append(k.cacheUpdatedWatchers[:i], k.cacheUpdatedWatchers[i+1:]...)
			return
		}
	}
}

func (k *kubeCoreCaches) updatedOccured() {
	k.cacheUpdatedWatchersMutex.Lock()
	defer k.cacheUpdatedWatchersMutex.Unlock()
	for _, cacheUpdated := range k.cacheUpdatedWatchers {
		select {
		case cacheUpdated <- struct{}{}:
		default:
		}
	}
}
