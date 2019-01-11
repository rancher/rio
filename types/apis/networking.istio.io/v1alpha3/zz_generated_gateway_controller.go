package v1alpha3

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
	GatewayGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "Gateway",
	}
	GatewayResource = metav1.APIResource{
		Name:         "gateways",
		SingularName: "gateway",
		Namespaced:   true,

		Kind: GatewayGroupVersionKind.Kind,
	}
)

func NewGateway(namespace, name string, obj Gateway) *Gateway {
	obj.APIVersion, obj.Kind = GatewayGroupVersionKind.ToAPIVersionAndKind()
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

type GatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Gateway
}

type GatewayHandlerFunc func(key string, obj *Gateway) (runtime.Object, error)

type GatewayChangeHandlerFunc func(obj *Gateway) (runtime.Object, error)

type GatewayLister interface {
	List(namespace string, selector labels.Selector) (ret []*Gateway, err error)
	Get(namespace, name string) (*Gateway, error)
}

type GatewayController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() GatewayLister
	AddHandler(ctx context.Context, name string, handler GatewayHandlerFunc)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, handler GatewayHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type GatewayInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*Gateway) (*Gateway, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*Gateway, error)
	Get(name string, opts metav1.GetOptions) (*Gateway, error)
	Update(*Gateway) (*Gateway, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*GatewayList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() GatewayController
	AddHandler(ctx context.Context, name string, sync GatewayHandlerFunc)
	AddLifecycle(ctx context.Context, name string, lifecycle GatewayLifecycle)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync GatewayHandlerFunc)
	AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle GatewayLifecycle)
}

type gatewayLister struct {
	controller *gatewayController
}

func (l *gatewayLister) List(namespace string, selector labels.Selector) (ret []*Gateway, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*Gateway))
	})
	return
}

func (l *gatewayLister) Get(namespace, name string) (*Gateway, error) {
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
			Group:    GatewayGroupVersionKind.Group,
			Resource: "gateway",
		}, key)
	}
	return obj.(*Gateway), nil
}

type gatewayController struct {
	controller.GenericController
}

func (c *gatewayController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *gatewayController) Lister() GatewayLister {
	return &gatewayLister{
		controller: c,
	}
}

func (c *gatewayController) AddHandler(ctx context.Context, name string, handler GatewayHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*Gateway); ok {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

func (c *gatewayController) AddClusterScopedHandler(ctx context.Context, name, cluster string, handler GatewayHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*Gateway); ok && controller.ObjectInCluster(cluster, obj) {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

type gatewayFactory struct {
}

func (c gatewayFactory) Object() runtime.Object {
	return &Gateway{}
}

func (c gatewayFactory) List() runtime.Object {
	return &GatewayList{}
}

func (s *gatewayClient) Controller() GatewayController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.gatewayControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(GatewayGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &gatewayController{
		GenericController: genericController,
	}

	s.client.gatewayControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type gatewayClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   GatewayController
}

func (s *gatewayClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *gatewayClient) Create(o *Gateway) (*Gateway, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*Gateway), err
}

func (s *gatewayClient) Get(name string, opts metav1.GetOptions) (*Gateway, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*Gateway), err
}

func (s *gatewayClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*Gateway, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*Gateway), err
}

func (s *gatewayClient) Update(o *Gateway) (*Gateway, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*Gateway), err
}

func (s *gatewayClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *gatewayClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *gatewayClient) List(opts metav1.ListOptions) (*GatewayList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*GatewayList), err
}

func (s *gatewayClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *gatewayClient) Patch(o *Gateway, patchType types.PatchType, data []byte, subresources ...string) (*Gateway, error) {
	obj, err := s.objectClient.Patch(o.Name, o, patchType, data, subresources...)
	return obj.(*Gateway), err
}

func (s *gatewayClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *gatewayClient) AddHandler(ctx context.Context, name string, sync GatewayHandlerFunc) {
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *gatewayClient) AddLifecycle(ctx context.Context, name string, lifecycle GatewayLifecycle) {
	sync := NewGatewayLifecycleAdapter(name, false, s, lifecycle)
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *gatewayClient) AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync GatewayHandlerFunc) {
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

func (s *gatewayClient) AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle GatewayLifecycle) {
	sync := NewGatewayLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

type GatewayIndexer func(obj *Gateway) ([]string, error)

type GatewayClientCache interface {
	Get(namespace, name string) (*Gateway, error)
	List(namespace string, selector labels.Selector) ([]*Gateway, error)

	Index(name string, indexer GatewayIndexer)
	GetIndexed(name, key string) ([]*Gateway, error)
}

type GatewayClient interface {
	Create(*Gateway) (*Gateway, error)
	Get(namespace, name string, opts metav1.GetOptions) (*Gateway, error)
	Update(*Gateway) (*Gateway, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	List(namespace string, opts metav1.ListOptions) (*GatewayList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)

	Cache() GatewayClientCache

	OnCreate(ctx context.Context, name string, sync GatewayChangeHandlerFunc)
	OnChange(ctx context.Context, name string, sync GatewayChangeHandlerFunc)
	OnRemove(ctx context.Context, name string, sync GatewayChangeHandlerFunc)
	Enqueue(namespace, name string)

	Generic() controller.GenericController
	ObjectClient() *objectclient.ObjectClient
	Interface() GatewayInterface
}

type gatewayClientCache struct {
	client *gatewayClient2
}

type gatewayClient2 struct {
	iface      GatewayInterface
	controller GatewayController
}

func (n *gatewayClient2) Interface() GatewayInterface {
	return n.iface
}

func (n *gatewayClient2) Generic() controller.GenericController {
	return n.iface.Controller().Generic()
}

func (n *gatewayClient2) ObjectClient() *objectclient.ObjectClient {
	return n.Interface().ObjectClient()
}

func (n *gatewayClient2) Enqueue(namespace, name string) {
	n.iface.Controller().Enqueue(namespace, name)
}

func (n *gatewayClient2) Create(obj *Gateway) (*Gateway, error) {
	return n.iface.Create(obj)
}

func (n *gatewayClient2) Get(namespace, name string, opts metav1.GetOptions) (*Gateway, error) {
	return n.iface.GetNamespaced(namespace, name, opts)
}

func (n *gatewayClient2) Update(obj *Gateway) (*Gateway, error) {
	return n.iface.Update(obj)
}

func (n *gatewayClient2) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	return n.iface.DeleteNamespaced(namespace, name, options)
}

func (n *gatewayClient2) List(namespace string, opts metav1.ListOptions) (*GatewayList, error) {
	return n.iface.List(opts)
}

func (n *gatewayClient2) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return n.iface.Watch(opts)
}

func (n *gatewayClientCache) Get(namespace, name string) (*Gateway, error) {
	return n.client.controller.Lister().Get(namespace, name)
}

func (n *gatewayClientCache) List(namespace string, selector labels.Selector) ([]*Gateway, error) {
	return n.client.controller.Lister().List(namespace, selector)
}

func (n *gatewayClient2) Cache() GatewayClientCache {
	n.loadController()
	return &gatewayClientCache{
		client: n,
	}
}

func (n *gatewayClient2) OnCreate(ctx context.Context, name string, sync GatewayChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-create", &gatewayLifecycleDelegate{create: sync})
}

func (n *gatewayClient2) OnChange(ctx context.Context, name string, sync GatewayChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-change", &gatewayLifecycleDelegate{update: sync})
}

func (n *gatewayClient2) OnRemove(ctx context.Context, name string, sync GatewayChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name, &gatewayLifecycleDelegate{remove: sync})
}

func (n *gatewayClientCache) Index(name string, indexer GatewayIndexer) {
	err := n.client.controller.Informer().GetIndexer().AddIndexers(map[string]cache.IndexFunc{
		name: func(obj interface{}) ([]string, error) {
			if v, ok := obj.(*Gateway); ok {
				return indexer(v)
			}
			return nil, nil
		},
	})

	if err != nil {
		panic(err)
	}
}

func (n *gatewayClientCache) GetIndexed(name, key string) ([]*Gateway, error) {
	var result []*Gateway
	objs, err := n.client.controller.Informer().GetIndexer().ByIndex(name, key)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		if v, ok := obj.(*Gateway); ok {
			result = append(result, v)
		}
	}

	return result, nil
}

func (n *gatewayClient2) loadController() {
	if n.controller == nil {
		n.controller = n.iface.Controller()
	}
}

type gatewayLifecycleDelegate struct {
	create GatewayChangeHandlerFunc
	update GatewayChangeHandlerFunc
	remove GatewayChangeHandlerFunc
}

func (n *gatewayLifecycleDelegate) HasCreate() bool {
	return n.create != nil
}

func (n *gatewayLifecycleDelegate) Create(obj *Gateway) (runtime.Object, error) {
	if n.create == nil {
		return obj, nil
	}
	return n.create(obj)
}

func (n *gatewayLifecycleDelegate) HasFinalize() bool {
	return n.remove != nil
}

func (n *gatewayLifecycleDelegate) Remove(obj *Gateway) (runtime.Object, error) {
	if n.remove == nil {
		return obj, nil
	}
	return n.remove(obj)
}

func (n *gatewayLifecycleDelegate) Updated(obj *Gateway) (runtime.Object, error) {
	if n.update == nil {
		return obj, nil
	}
	return n.update(obj)
}
