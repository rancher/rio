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
	StackGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "Stack",
	}
	StackResource = metav1.APIResource{
		Name:         "stacks",
		SingularName: "stack",
		Namespaced:   true,

		Kind: StackGroupVersionKind.Kind,
	}
)

func NewStack(namespace, name string, obj Stack) *Stack {
	obj.APIVersion, obj.Kind = StackGroupVersionKind.ToAPIVersionAndKind()
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

type StackList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Stack
}

type StackHandlerFunc func(key string, obj *Stack) (runtime.Object, error)

type StackChangeHandlerFunc func(obj *Stack) (runtime.Object, error)

type StackLister interface {
	List(namespace string, selector labels.Selector) (ret []*Stack, err error)
	Get(namespace, name string) (*Stack, error)
}

type StackController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() StackLister
	AddHandler(ctx context.Context, name string, handler StackHandlerFunc)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, handler StackHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type StackInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*Stack) (*Stack, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*Stack, error)
	Get(name string, opts metav1.GetOptions) (*Stack, error)
	Update(*Stack) (*Stack, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*StackList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() StackController
	AddHandler(ctx context.Context, name string, sync StackHandlerFunc)
	AddLifecycle(ctx context.Context, name string, lifecycle StackLifecycle)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync StackHandlerFunc)
	AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle StackLifecycle)
}

type stackLister struct {
	controller *stackController
}

func (l *stackLister) List(namespace string, selector labels.Selector) (ret []*Stack, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*Stack))
	})
	return
}

func (l *stackLister) Get(namespace, name string) (*Stack, error) {
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
			Group:    StackGroupVersionKind.Group,
			Resource: "stack",
		}, key)
	}
	return obj.(*Stack), nil
}

type stackController struct {
	controller.GenericController
}

func (c *stackController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *stackController) Lister() StackLister {
	return &stackLister{
		controller: c,
	}
}

func (c *stackController) AddHandler(ctx context.Context, name string, handler StackHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*Stack); ok {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

func (c *stackController) AddClusterScopedHandler(ctx context.Context, name, cluster string, handler StackHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*Stack); ok && controller.ObjectInCluster(cluster, obj) {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

type stackFactory struct {
}

func (c stackFactory) Object() runtime.Object {
	return &Stack{}
}

func (c stackFactory) List() runtime.Object {
	return &StackList{}
}

func (s *stackClient) Controller() StackController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.stackControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(StackGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &stackController{
		GenericController: genericController,
	}

	s.client.stackControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type stackClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   StackController
}

func (s *stackClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *stackClient) Create(o *Stack) (*Stack, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*Stack), err
}

func (s *stackClient) Get(name string, opts metav1.GetOptions) (*Stack, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*Stack), err
}

func (s *stackClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*Stack, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*Stack), err
}

func (s *stackClient) Update(o *Stack) (*Stack, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*Stack), err
}

func (s *stackClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *stackClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *stackClient) List(opts metav1.ListOptions) (*StackList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*StackList), err
}

func (s *stackClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *stackClient) Patch(o *Stack, patchType types.PatchType, data []byte, subresources ...string) (*Stack, error) {
	obj, err := s.objectClient.Patch(o.Name, o, patchType, data, subresources...)
	return obj.(*Stack), err
}

func (s *stackClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *stackClient) AddHandler(ctx context.Context, name string, sync StackHandlerFunc) {
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *stackClient) AddLifecycle(ctx context.Context, name string, lifecycle StackLifecycle) {
	sync := NewStackLifecycleAdapter(name, false, s, lifecycle)
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *stackClient) AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync StackHandlerFunc) {
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

func (s *stackClient) AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle StackLifecycle) {
	sync := NewStackLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

type StackIndexer func(obj *Stack) ([]string, error)

type StackClientCache interface {
	Get(namespace, name string) (*Stack, error)
	List(namespace string, selector labels.Selector) ([]*Stack, error)

	Index(name string, indexer StackIndexer)
	GetIndexed(name, key string) ([]*Stack, error)
}

type StackClient interface {
	Create(*Stack) (*Stack, error)
	Get(namespace, name string, opts metav1.GetOptions) (*Stack, error)
	Update(*Stack) (*Stack, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	List(namespace string, opts metav1.ListOptions) (*StackList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)

	Cache() StackClientCache

	OnCreate(ctx context.Context, name string, sync StackChangeHandlerFunc)
	OnChange(ctx context.Context, name string, sync StackChangeHandlerFunc)
	OnRemove(ctx context.Context, name string, sync StackChangeHandlerFunc)
	Enqueue(namespace, name string)

	Generic() controller.GenericController
	ObjectClient() *objectclient.ObjectClient
	Interface() StackInterface
}

type stackClientCache struct {
	client *stackClient2
}

type stackClient2 struct {
	iface      StackInterface
	controller StackController
}

func (n *stackClient2) Interface() StackInterface {
	return n.iface
}

func (n *stackClient2) Generic() controller.GenericController {
	return n.iface.Controller().Generic()
}

func (n *stackClient2) ObjectClient() *objectclient.ObjectClient {
	return n.Interface().ObjectClient()
}

func (n *stackClient2) Enqueue(namespace, name string) {
	n.iface.Controller().Enqueue(namespace, name)
}

func (n *stackClient2) Create(obj *Stack) (*Stack, error) {
	return n.iface.Create(obj)
}

func (n *stackClient2) Get(namespace, name string, opts metav1.GetOptions) (*Stack, error) {
	return n.iface.GetNamespaced(namespace, name, opts)
}

func (n *stackClient2) Update(obj *Stack) (*Stack, error) {
	return n.iface.Update(obj)
}

func (n *stackClient2) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	return n.iface.DeleteNamespaced(namespace, name, options)
}

func (n *stackClient2) List(namespace string, opts metav1.ListOptions) (*StackList, error) {
	return n.iface.List(opts)
}

func (n *stackClient2) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return n.iface.Watch(opts)
}

func (n *stackClientCache) Get(namespace, name string) (*Stack, error) {
	return n.client.controller.Lister().Get(namespace, name)
}

func (n *stackClientCache) List(namespace string, selector labels.Selector) ([]*Stack, error) {
	return n.client.controller.Lister().List(namespace, selector)
}

func (n *stackClient2) Cache() StackClientCache {
	n.loadController()
	return &stackClientCache{
		client: n,
	}
}

func (n *stackClient2) OnCreate(ctx context.Context, name string, sync StackChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-create", &stackLifecycleDelegate{create: sync})
}

func (n *stackClient2) OnChange(ctx context.Context, name string, sync StackChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-change", &stackLifecycleDelegate{update: sync})
}

func (n *stackClient2) OnRemove(ctx context.Context, name string, sync StackChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name, &stackLifecycleDelegate{remove: sync})
}

func (n *stackClientCache) Index(name string, indexer StackIndexer) {
	err := n.client.controller.Informer().GetIndexer().AddIndexers(map[string]cache.IndexFunc{
		name: func(obj interface{}) ([]string, error) {
			if v, ok := obj.(*Stack); ok {
				return indexer(v)
			}
			return nil, nil
		},
	})

	if err != nil {
		panic(err)
	}
}

func (n *stackClientCache) GetIndexed(name, key string) ([]*Stack, error) {
	var result []*Stack
	objs, err := n.client.controller.Informer().GetIndexer().ByIndex(name, key)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		if v, ok := obj.(*Stack); ok {
			result = append(result, v)
		}
	}

	return result, nil
}

func (n *stackClient2) loadController() {
	if n.controller == nil {
		n.controller = n.iface.Controller()
	}
}

type stackLifecycleDelegate struct {
	create StackChangeHandlerFunc
	update StackChangeHandlerFunc
	remove StackChangeHandlerFunc
}

func (n *stackLifecycleDelegate) HasCreate() bool {
	return n.create != nil
}

func (n *stackLifecycleDelegate) Create(obj *Stack) (runtime.Object, error) {
	if n.create == nil {
		return obj, nil
	}
	return n.create(obj)
}

func (n *stackLifecycleDelegate) HasFinalize() bool {
	return n.remove != nil
}

func (n *stackLifecycleDelegate) Remove(obj *Stack) (runtime.Object, error) {
	if n.remove == nil {
		return obj, nil
	}
	return n.remove(obj)
}

func (n *stackLifecycleDelegate) Updated(obj *Stack) (runtime.Object, error) {
	if n.update == nil {
		return obj, nil
	}
	return n.update(obj)
}
