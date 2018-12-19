package v1

import (
	"context"

	"github.com/rancher/norman/controller"
	"github.com/rancher/norman/objectclient"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

var (
	RouteSetGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "RouteSet",
	}
	RouteSetResource = metav1.APIResource{
		Name:         "routesets",
		SingularName: "routeset",
		Namespaced:   true,

		Kind: RouteSetGroupVersionKind.Kind,
	}
)

func NewRouteSet(namespace, name string, obj RouteSet) *RouteSet {
	obj.APIVersion, obj.Kind = RouteSetGroupVersionKind.ToAPIVersionAndKind()
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

type RouteSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RouteSet
}

type RouteSetHandlerFunc func(key string, obj *RouteSet) (runtime.Object, error)

type RouteSetChangeHandlerFunc func(obj *RouteSet) (runtime.Object, error)

type RouteSetLister interface {
	List(namespace string, selector labels.Selector) (ret []*RouteSet, err error)
	Get(namespace, name string) (*RouteSet, error)
}

type RouteSetController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() RouteSetLister
	AddHandler(ctx context.Context, name string, handler RouteSetHandlerFunc)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, handler RouteSetHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type RouteSetInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*RouteSet) (*RouteSet, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*RouteSet, error)
	Get(name string, opts metav1.GetOptions) (*RouteSet, error)
	Update(*RouteSet) (*RouteSet, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*RouteSetList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() RouteSetController
	AddHandler(ctx context.Context, name string, sync RouteSetHandlerFunc)
	AddLifecycle(ctx context.Context, name string, lifecycle RouteSetLifecycle)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync RouteSetHandlerFunc)
	AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle RouteSetLifecycle)
}

type routeSetLister struct {
	controller *routeSetController
}

func (l *routeSetLister) List(namespace string, selector labels.Selector) (ret []*RouteSet, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*RouteSet))
	})
	return
}

func (l *routeSetLister) Get(namespace, name string) (*RouteSet, error) {
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
			Group:    RouteSetGroupVersionKind.Group,
			Resource: "routeSet",
		}, key)
	}
	return obj.(*RouteSet), nil
}

type routeSetController struct {
	controller.GenericController
}

func (c *routeSetController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *routeSetController) Lister() RouteSetLister {
	return &routeSetLister{
		controller: c,
	}
}

func (c *routeSetController) AddHandler(ctx context.Context, name string, handler RouteSetHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*RouteSet); ok {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

func (c *routeSetController) AddClusterScopedHandler(ctx context.Context, name, cluster string, handler RouteSetHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*RouteSet); ok && controller.ObjectInCluster(cluster, obj) {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

type routeSetFactory struct {
}

func (c routeSetFactory) Object() runtime.Object {
	return &RouteSet{}
}

func (c routeSetFactory) List() runtime.Object {
	return &RouteSetList{}
}

func (s *routeSetClient) Controller() RouteSetController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.routeSetControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(RouteSetGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &routeSetController{
		GenericController: genericController,
	}

	s.client.routeSetControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type routeSetClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   RouteSetController
}

func (s *routeSetClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *routeSetClient) Create(o *RouteSet) (*RouteSet, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*RouteSet), err
}

func (s *routeSetClient) Get(name string, opts metav1.GetOptions) (*RouteSet, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*RouteSet), err
}

func (s *routeSetClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*RouteSet, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*RouteSet), err
}

func (s *routeSetClient) Update(o *RouteSet) (*RouteSet, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*RouteSet), err
}

func (s *routeSetClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *routeSetClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *routeSetClient) List(opts metav1.ListOptions) (*RouteSetList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*RouteSetList), err
}

func (s *routeSetClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *routeSetClient) Patch(o *RouteSet, patchType types.PatchType, data []byte, subresources ...string) (*RouteSet, error) {
	obj, err := s.objectClient.Patch(o.Name, o, patchType, data, subresources...)
	return obj.(*RouteSet), err
}

func (s *routeSetClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *routeSetClient) AddHandler(ctx context.Context, name string, sync RouteSetHandlerFunc) {
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *routeSetClient) AddLifecycle(ctx context.Context, name string, lifecycle RouteSetLifecycle) {
	sync := NewRouteSetLifecycleAdapter(name, false, s, lifecycle)
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *routeSetClient) AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync RouteSetHandlerFunc) {
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

func (s *routeSetClient) AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle RouteSetLifecycle) {
	sync := NewRouteSetLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

type RouteSetIndexer func(obj *RouteSet) ([]string, error)

type RouteSetClientCache interface {
	Get(namespace, name string) (*RouteSet, error)
	List(namespace string, selector labels.Selector) ([]*RouteSet, error)

	Index(name string, indexer RouteSetIndexer)
	GetIndexed(name, key string) ([]*RouteSet, error)
}

type RouteSetClient interface {
	Create(*RouteSet) (*RouteSet, error)
	Get(namespace, name string, opts metav1.GetOptions) (*RouteSet, error)
	Update(*RouteSet) (*RouteSet, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	List(namespace string, opts metav1.ListOptions) (*RouteSetList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)

	Cache() RouteSetClientCache

	OnCreate(ctx context.Context, name string, sync RouteSetChangeHandlerFunc)
	OnChange(ctx context.Context, name string, sync RouteSetChangeHandlerFunc)
	OnRemove(ctx context.Context, name string, sync RouteSetChangeHandlerFunc)
	Enqueue(namespace, name string)

	Generic() controller.GenericController
	ObjectClient() *objectclient.ObjectClient
	Interface() RouteSetInterface
}

type routeSetClientCache struct {
	client *routeSetClient2
}

type routeSetClient2 struct {
	iface      RouteSetInterface
	controller RouteSetController
}

func (n *routeSetClient2) Interface() RouteSetInterface {
	return n.iface
}

func (n *routeSetClient2) Generic() controller.GenericController {
	return n.iface.Controller().Generic()
}

func (n *routeSetClient2) ObjectClient() *objectclient.ObjectClient {
	return n.Interface().ObjectClient()
}

func (n *routeSetClient2) Enqueue(namespace, name string) {
	n.iface.Controller().Enqueue(namespace, name)
}

func (n *routeSetClient2) Create(obj *RouteSet) (*RouteSet, error) {
	return n.iface.Create(obj)
}

func (n *routeSetClient2) Get(namespace, name string, opts metav1.GetOptions) (*RouteSet, error) {
	return n.iface.GetNamespaced(namespace, name, opts)
}

func (n *routeSetClient2) Update(obj *RouteSet) (*RouteSet, error) {
	return n.iface.Update(obj)
}

func (n *routeSetClient2) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	return n.iface.DeleteNamespaced(namespace, name, options)
}

func (n *routeSetClient2) List(namespace string, opts metav1.ListOptions) (*RouteSetList, error) {
	return n.iface.List(opts)
}

func (n *routeSetClient2) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return n.iface.Watch(opts)
}

func (n *routeSetClientCache) Get(namespace, name string) (*RouteSet, error) {
	return n.client.controller.Lister().Get(namespace, name)
}

func (n *routeSetClientCache) List(namespace string, selector labels.Selector) ([]*RouteSet, error) {
	return n.client.controller.Lister().List(namespace, selector)
}

func (n *routeSetClient2) Cache() RouteSetClientCache {
	n.loadController()
	return &routeSetClientCache{
		client: n,
	}
}

func (n *routeSetClient2) OnCreate(ctx context.Context, name string, sync RouteSetChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-create", &routeSetLifecycleDelegate{create: sync})
}

func (n *routeSetClient2) OnChange(ctx context.Context, name string, sync RouteSetChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-change", &routeSetLifecycleDelegate{update: sync})
}

func (n *routeSetClient2) OnRemove(ctx context.Context, name string, sync RouteSetChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name, &routeSetLifecycleDelegate{remove: sync})
}

func (n *routeSetClientCache) Index(name string, indexer RouteSetIndexer) {
	err := n.client.controller.Informer().GetIndexer().AddIndexers(map[string]cache.IndexFunc{
		name: func(obj interface{}) ([]string, error) {
			if v, ok := obj.(*RouteSet); ok {
				return indexer(v)
			}
			return nil, nil
		},
	})

	if err != nil {
		panic(err)
	}
}

func (n *routeSetClientCache) GetIndexed(name, key string) ([]*RouteSet, error) {
	var result []*RouteSet
	objs, err := n.client.controller.Informer().GetIndexer().ByIndex(name, key)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		if v, ok := obj.(*RouteSet); ok {
			result = append(result, v)
		}
	}

	return result, nil
}

func (n *routeSetClient2) loadController() {
	if n.controller == nil {
		n.controller = n.iface.Controller()
	}
}

type routeSetLifecycleDelegate struct {
	create RouteSetChangeHandlerFunc
	update RouteSetChangeHandlerFunc
	remove RouteSetChangeHandlerFunc
}

func (n *routeSetLifecycleDelegate) HasCreate() bool {
	return n.create != nil
}

func (n *routeSetLifecycleDelegate) Create(obj *RouteSet) (runtime.Object, error) {
	if n.create == nil {
		return obj, nil
	}
	return n.create(obj)
}

func (n *routeSetLifecycleDelegate) HasFinalize() bool {
	return n.remove != nil
}

func (n *routeSetLifecycleDelegate) Remove(obj *RouteSet) (runtime.Object, error) {
	if n.remove == nil {
		return obj, nil
	}
	return n.remove(obj)
}

func (n *routeSetLifecycleDelegate) Updated(obj *RouteSet) (runtime.Object, error) {
	if n.update == nil {
		return obj, nil
	}
	return n.update(obj)
}
