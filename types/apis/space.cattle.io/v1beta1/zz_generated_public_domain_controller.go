package v1beta1

import (
	"context"

	"github.com/rancher/norman/controller"
	"github.com/rancher/norman/objectclient"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

var (
	PublicDomainGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "PublicDomain",
	}
	PublicDomainResource = metav1.APIResource{
		Name:         "publicdomains",
		SingularName: "publicdomain",
		Namespaced:   true,

		Kind: PublicDomainGroupVersionKind.Kind,
	}
)

type PublicDomainList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PublicDomain
}

type PublicDomainHandlerFunc func(key string, obj *PublicDomain) (runtime.Object, error)

type PublicDomainChangeHandlerFunc func(obj *PublicDomain) (runtime.Object, error)

type PublicDomainLister interface {
	List(namespace string, selector labels.Selector) (ret []*PublicDomain, err error)
	Get(namespace, name string) (*PublicDomain, error)
}

type PublicDomainController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() PublicDomainLister
	AddHandler(ctx context.Context, name string, handler PublicDomainHandlerFunc)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, handler PublicDomainHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type PublicDomainInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*PublicDomain) (*PublicDomain, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*PublicDomain, error)
	Get(name string, opts metav1.GetOptions) (*PublicDomain, error)
	Update(*PublicDomain) (*PublicDomain, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*PublicDomainList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() PublicDomainController
	AddHandler(ctx context.Context, name string, sync PublicDomainHandlerFunc)
	AddLifecycle(ctx context.Context, name string, lifecycle PublicDomainLifecycle)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync PublicDomainHandlerFunc)
	AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle PublicDomainLifecycle)
}

type publicDomainLister struct {
	controller *publicDomainController
}

func (l *publicDomainLister) List(namespace string, selector labels.Selector) (ret []*PublicDomain, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*PublicDomain))
	})
	return
}

func (l *publicDomainLister) Get(namespace, name string) (*PublicDomain, error) {
	var key string
	if namespace != "" {
		key = namespace + "/" + name
	} else {
		key = name
	}
	obj, exists, err := l.controller.Informer().GetIndexer().GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(schema.GroupResource{
			Group:    PublicDomainGroupVersionKind.Group,
			Resource: "publicDomain",
		}, key)
	}
	return obj.(*PublicDomain), nil
}

type publicDomainController struct {
	controller.GenericController
}

func (c *publicDomainController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *publicDomainController) Lister() PublicDomainLister {
	return &publicDomainLister{
		controller: c,
	}
}

func (c *publicDomainController) AddHandler(ctx context.Context, name string, handler PublicDomainHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*PublicDomain); ok {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

func (c *publicDomainController) AddClusterScopedHandler(ctx context.Context, name, cluster string, handler PublicDomainHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*PublicDomain); ok && controller.ObjectInCluster(cluster, obj) {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

type publicDomainFactory struct {
}

func (c publicDomainFactory) Object() runtime.Object {
	return &PublicDomain{}
}

func (c publicDomainFactory) List() runtime.Object {
	return &PublicDomainList{}
}

func (s *publicDomainClient) Controller() PublicDomainController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.publicDomainControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(PublicDomainGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &publicDomainController{
		GenericController: genericController,
	}

	s.client.publicDomainControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type publicDomainClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   PublicDomainController
}

func (s *publicDomainClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *publicDomainClient) Create(o *PublicDomain) (*PublicDomain, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*PublicDomain), err
}

func (s *publicDomainClient) Get(name string, opts metav1.GetOptions) (*PublicDomain, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*PublicDomain), err
}

func (s *publicDomainClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*PublicDomain, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*PublicDomain), err
}

func (s *publicDomainClient) Update(o *PublicDomain) (*PublicDomain, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*PublicDomain), err
}

func (s *publicDomainClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *publicDomainClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *publicDomainClient) List(opts metav1.ListOptions) (*PublicDomainList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*PublicDomainList), err
}

func (s *publicDomainClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *publicDomainClient) Patch(o *PublicDomain, data []byte, subresources ...string) (*PublicDomain, error) {
	obj, err := s.objectClient.Patch(o.Name, o, data, subresources...)
	return obj.(*PublicDomain), err
}

func (s *publicDomainClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *publicDomainClient) AddHandler(ctx context.Context, name string, sync PublicDomainHandlerFunc) {
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *publicDomainClient) AddLifecycle(ctx context.Context, name string, lifecycle PublicDomainLifecycle) {
	sync := NewPublicDomainLifecycleAdapter(name, false, s, lifecycle)
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *publicDomainClient) AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync PublicDomainHandlerFunc) {
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

func (s *publicDomainClient) AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle PublicDomainLifecycle) {
	sync := NewPublicDomainLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

type PublicDomainIndexer func(obj *PublicDomain) ([]string, error)

type PublicDomainClientCache interface {
	Get(namespace, name string) (*PublicDomain, error)
	List(namespace string, selector labels.Selector) ([]*PublicDomain, error)

	Index(name string, indexer PublicDomainIndexer)
	GetIndexed(name, key string) ([]*PublicDomain, error)
}

type PublicDomainClient interface {
	Create(*PublicDomain) (*PublicDomain, error)
	Get(namespace, name string, opts metav1.GetOptions) (*PublicDomain, error)
	Update(*PublicDomain) (*PublicDomain, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	List(namespace string, opts metav1.ListOptions) (*PublicDomainList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)

	Cache() PublicDomainClientCache

	OnCreate(ctx context.Context, name string, sync PublicDomainChangeHandlerFunc)
	OnChange(ctx context.Context, name string, sync PublicDomainChangeHandlerFunc)
	OnRemove(ctx context.Context, name string, sync PublicDomainChangeHandlerFunc)
	Enqueue(namespace, name string)

	Generic() controller.GenericController
	Interface() PublicDomainInterface
}

type publicDomainClientCache struct {
	client *publicDomainClient2
}

type publicDomainClient2 struct {
	iface      PublicDomainInterface
	controller PublicDomainController
}

func (n *publicDomainClient2) Interface() PublicDomainInterface {
	return n.iface
}

func (n *publicDomainClient2) Generic() controller.GenericController {
	return n.iface.Controller().Generic()
}

func (n *publicDomainClient2) Enqueue(namespace, name string) {
	n.iface.Controller().Enqueue(namespace, name)
}

func (n *publicDomainClient2) Create(obj *PublicDomain) (*PublicDomain, error) {
	return n.iface.Create(obj)
}

func (n *publicDomainClient2) Get(namespace, name string, opts metav1.GetOptions) (*PublicDomain, error) {
	return n.iface.GetNamespaced(namespace, name, opts)
}

func (n *publicDomainClient2) Update(obj *PublicDomain) (*PublicDomain, error) {
	return n.iface.Update(obj)
}

func (n *publicDomainClient2) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	return n.iface.DeleteNamespaced(namespace, name, options)
}

func (n *publicDomainClient2) List(namespace string, opts metav1.ListOptions) (*PublicDomainList, error) {
	return n.iface.List(opts)
}

func (n *publicDomainClient2) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return n.iface.Watch(opts)
}

func (n *publicDomainClientCache) Get(namespace, name string) (*PublicDomain, error) {
	return n.client.controller.Lister().Get(namespace, name)
}

func (n *publicDomainClientCache) List(namespace string, selector labels.Selector) ([]*PublicDomain, error) {
	return n.client.controller.Lister().List(namespace, selector)
}

func (n *publicDomainClient2) Cache() PublicDomainClientCache {
	n.loadController()
	return &publicDomainClientCache{
		client: n,
	}
}

func (n *publicDomainClient2) OnCreate(ctx context.Context, name string, sync PublicDomainChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-create", &publicDomainLifecycleDelegate{create: sync})
}

func (n *publicDomainClient2) OnChange(ctx context.Context, name string, sync PublicDomainChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-change", &publicDomainLifecycleDelegate{update: sync})
}

func (n *publicDomainClient2) OnRemove(ctx context.Context, name string, sync PublicDomainChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name, &publicDomainLifecycleDelegate{remove: sync})
}

func (n *publicDomainClientCache) Index(name string, indexer PublicDomainIndexer) {
	err := n.client.controller.Informer().GetIndexer().AddIndexers(map[string]cache.IndexFunc{
		name: func(obj interface{}) ([]string, error) {
			if v, ok := obj.(*PublicDomain); ok {
				return indexer(v)
			}
			return nil, nil
		},
	})

	if err != nil {
		panic(err)
	}
}

func (n *publicDomainClientCache) GetIndexed(name, key string) ([]*PublicDomain, error) {
	var result []*PublicDomain
	objs, err := n.client.controller.Informer().GetIndexer().ByIndex(name, key)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		if v, ok := obj.(*PublicDomain); ok {
			result = append(result, v)
		}
	}

	return result, nil
}

func (n *publicDomainClient2) loadController() {
	if n.controller == nil {
		n.controller = n.iface.Controller()
	}
}

type publicDomainLifecycleDelegate struct {
	create PublicDomainChangeHandlerFunc
	update PublicDomainChangeHandlerFunc
	remove PublicDomainChangeHandlerFunc
}

func (n *publicDomainLifecycleDelegate) HasCreate() bool {
	return n.create != nil
}

func (n *publicDomainLifecycleDelegate) Create(obj *PublicDomain) (runtime.Object, error) {
	if n.create == nil {
		return obj, nil
	}
	return n.create(obj)
}

func (n *publicDomainLifecycleDelegate) HasFinalize() bool {
	return n.remove != nil
}

func (n *publicDomainLifecycleDelegate) Remove(obj *PublicDomain) (runtime.Object, error) {
	if n.remove == nil {
		return obj, nil
	}
	return n.remove(obj)
}

func (n *publicDomainLifecycleDelegate) Updated(obj *PublicDomain) (runtime.Object, error) {
	if n.update == nil {
		return obj, nil
	}
	return n.update(obj)
}
