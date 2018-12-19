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
	ServiceEntryGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "ServiceEntry",
	}
	ServiceEntryResource = metav1.APIResource{
		Name:         "serviceentries",
		SingularName: "serviceentry",
		Namespaced:   true,

		Kind: ServiceEntryGroupVersionKind.Kind,
	}
)

func NewServiceEntry(namespace, name string, obj ServiceEntry) *ServiceEntry {
	obj.APIVersion, obj.Kind = ServiceEntryGroupVersionKind.ToAPIVersionAndKind()
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

type ServiceEntryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceEntry
}

type ServiceEntryHandlerFunc func(key string, obj *ServiceEntry) (runtime.Object, error)

type ServiceEntryChangeHandlerFunc func(obj *ServiceEntry) (runtime.Object, error)

type ServiceEntryLister interface {
	List(namespace string, selector labels.Selector) (ret []*ServiceEntry, err error)
	Get(namespace, name string) (*ServiceEntry, error)
}

type ServiceEntryController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() ServiceEntryLister
	AddHandler(ctx context.Context, name string, handler ServiceEntryHandlerFunc)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, handler ServiceEntryHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type ServiceEntryInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*ServiceEntry) (*ServiceEntry, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*ServiceEntry, error)
	Get(name string, opts metav1.GetOptions) (*ServiceEntry, error)
	Update(*ServiceEntry) (*ServiceEntry, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*ServiceEntryList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() ServiceEntryController
	AddHandler(ctx context.Context, name string, sync ServiceEntryHandlerFunc)
	AddLifecycle(ctx context.Context, name string, lifecycle ServiceEntryLifecycle)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync ServiceEntryHandlerFunc)
	AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle ServiceEntryLifecycle)
}

type serviceEntryLister struct {
	controller *serviceEntryController
}

func (l *serviceEntryLister) List(namespace string, selector labels.Selector) (ret []*ServiceEntry, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*ServiceEntry))
	})
	return
}

func (l *serviceEntryLister) Get(namespace, name string) (*ServiceEntry, error) {
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
			Group:    ServiceEntryGroupVersionKind.Group,
			Resource: "serviceEntry",
		}, key)
	}
	return obj.(*ServiceEntry), nil
}

type serviceEntryController struct {
	controller.GenericController
}

func (c *serviceEntryController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *serviceEntryController) Lister() ServiceEntryLister {
	return &serviceEntryLister{
		controller: c,
	}
}

func (c *serviceEntryController) AddHandler(ctx context.Context, name string, handler ServiceEntryHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*ServiceEntry); ok {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

func (c *serviceEntryController) AddClusterScopedHandler(ctx context.Context, name, cluster string, handler ServiceEntryHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*ServiceEntry); ok && controller.ObjectInCluster(cluster, obj) {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

type serviceEntryFactory struct {
}

func (c serviceEntryFactory) Object() runtime.Object {
	return &ServiceEntry{}
}

func (c serviceEntryFactory) List() runtime.Object {
	return &ServiceEntryList{}
}

func (s *serviceEntryClient) Controller() ServiceEntryController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.serviceEntryControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(ServiceEntryGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &serviceEntryController{
		GenericController: genericController,
	}

	s.client.serviceEntryControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type serviceEntryClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   ServiceEntryController
}

func (s *serviceEntryClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *serviceEntryClient) Create(o *ServiceEntry) (*ServiceEntry, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*ServiceEntry), err
}

func (s *serviceEntryClient) Get(name string, opts metav1.GetOptions) (*ServiceEntry, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*ServiceEntry), err
}

func (s *serviceEntryClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*ServiceEntry, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*ServiceEntry), err
}

func (s *serviceEntryClient) Update(o *ServiceEntry) (*ServiceEntry, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*ServiceEntry), err
}

func (s *serviceEntryClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *serviceEntryClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *serviceEntryClient) List(opts metav1.ListOptions) (*ServiceEntryList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*ServiceEntryList), err
}

func (s *serviceEntryClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *serviceEntryClient) Patch(o *ServiceEntry, patchType types.PatchType, data []byte, subresources ...string) (*ServiceEntry, error) {
	obj, err := s.objectClient.Patch(o.Name, o, patchType, data, subresources...)
	return obj.(*ServiceEntry), err
}

func (s *serviceEntryClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *serviceEntryClient) AddHandler(ctx context.Context, name string, sync ServiceEntryHandlerFunc) {
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *serviceEntryClient) AddLifecycle(ctx context.Context, name string, lifecycle ServiceEntryLifecycle) {
	sync := NewServiceEntryLifecycleAdapter(name, false, s, lifecycle)
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *serviceEntryClient) AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync ServiceEntryHandlerFunc) {
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

func (s *serviceEntryClient) AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle ServiceEntryLifecycle) {
	sync := NewServiceEntryLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

type ServiceEntryIndexer func(obj *ServiceEntry) ([]string, error)

type ServiceEntryClientCache interface {
	Get(namespace, name string) (*ServiceEntry, error)
	List(namespace string, selector labels.Selector) ([]*ServiceEntry, error)

	Index(name string, indexer ServiceEntryIndexer)
	GetIndexed(name, key string) ([]*ServiceEntry, error)
}

type ServiceEntryClient interface {
	Create(*ServiceEntry) (*ServiceEntry, error)
	Get(namespace, name string, opts metav1.GetOptions) (*ServiceEntry, error)
	Update(*ServiceEntry) (*ServiceEntry, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	List(namespace string, opts metav1.ListOptions) (*ServiceEntryList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)

	Cache() ServiceEntryClientCache

	OnCreate(ctx context.Context, name string, sync ServiceEntryChangeHandlerFunc)
	OnChange(ctx context.Context, name string, sync ServiceEntryChangeHandlerFunc)
	OnRemove(ctx context.Context, name string, sync ServiceEntryChangeHandlerFunc)
	Enqueue(namespace, name string)

	Generic() controller.GenericController
	ObjectClient() *objectclient.ObjectClient
	Interface() ServiceEntryInterface
}

type serviceEntryClientCache struct {
	client *serviceEntryClient2
}

type serviceEntryClient2 struct {
	iface      ServiceEntryInterface
	controller ServiceEntryController
}

func (n *serviceEntryClient2) Interface() ServiceEntryInterface {
	return n.iface
}

func (n *serviceEntryClient2) Generic() controller.GenericController {
	return n.iface.Controller().Generic()
}

func (n *serviceEntryClient2) ObjectClient() *objectclient.ObjectClient {
	return n.Interface().ObjectClient()
}

func (n *serviceEntryClient2) Enqueue(namespace, name string) {
	n.iface.Controller().Enqueue(namespace, name)
}

func (n *serviceEntryClient2) Create(obj *ServiceEntry) (*ServiceEntry, error) {
	return n.iface.Create(obj)
}

func (n *serviceEntryClient2) Get(namespace, name string, opts metav1.GetOptions) (*ServiceEntry, error) {
	return n.iface.GetNamespaced(namespace, name, opts)
}

func (n *serviceEntryClient2) Update(obj *ServiceEntry) (*ServiceEntry, error) {
	return n.iface.Update(obj)
}

func (n *serviceEntryClient2) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	return n.iface.DeleteNamespaced(namespace, name, options)
}

func (n *serviceEntryClient2) List(namespace string, opts metav1.ListOptions) (*ServiceEntryList, error) {
	return n.iface.List(opts)
}

func (n *serviceEntryClient2) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return n.iface.Watch(opts)
}

func (n *serviceEntryClientCache) Get(namespace, name string) (*ServiceEntry, error) {
	return n.client.controller.Lister().Get(namespace, name)
}

func (n *serviceEntryClientCache) List(namespace string, selector labels.Selector) ([]*ServiceEntry, error) {
	return n.client.controller.Lister().List(namespace, selector)
}

func (n *serviceEntryClient2) Cache() ServiceEntryClientCache {
	n.loadController()
	return &serviceEntryClientCache{
		client: n,
	}
}

func (n *serviceEntryClient2) OnCreate(ctx context.Context, name string, sync ServiceEntryChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-create", &serviceEntryLifecycleDelegate{create: sync})
}

func (n *serviceEntryClient2) OnChange(ctx context.Context, name string, sync ServiceEntryChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-change", &serviceEntryLifecycleDelegate{update: sync})
}

func (n *serviceEntryClient2) OnRemove(ctx context.Context, name string, sync ServiceEntryChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name, &serviceEntryLifecycleDelegate{remove: sync})
}

func (n *serviceEntryClientCache) Index(name string, indexer ServiceEntryIndexer) {
	err := n.client.controller.Informer().GetIndexer().AddIndexers(map[string]cache.IndexFunc{
		name: func(obj interface{}) ([]string, error) {
			if v, ok := obj.(*ServiceEntry); ok {
				return indexer(v)
			}
			return nil, nil
		},
	})

	if err != nil {
		panic(err)
	}
}

func (n *serviceEntryClientCache) GetIndexed(name, key string) ([]*ServiceEntry, error) {
	var result []*ServiceEntry
	objs, err := n.client.controller.Informer().GetIndexer().ByIndex(name, key)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		if v, ok := obj.(*ServiceEntry); ok {
			result = append(result, v)
		}
	}

	return result, nil
}

func (n *serviceEntryClient2) loadController() {
	if n.controller == nil {
		n.controller = n.iface.Controller()
	}
}

type serviceEntryLifecycleDelegate struct {
	create ServiceEntryChangeHandlerFunc
	update ServiceEntryChangeHandlerFunc
	remove ServiceEntryChangeHandlerFunc
}

func (n *serviceEntryLifecycleDelegate) HasCreate() bool {
	return n.create != nil
}

func (n *serviceEntryLifecycleDelegate) Create(obj *ServiceEntry) (runtime.Object, error) {
	if n.create == nil {
		return obj, nil
	}
	return n.create(obj)
}

func (n *serviceEntryLifecycleDelegate) HasFinalize() bool {
	return n.remove != nil
}

func (n *serviceEntryLifecycleDelegate) Remove(obj *ServiceEntry) (runtime.Object, error) {
	if n.remove == nil {
		return obj, nil
	}
	return n.remove(obj)
}

func (n *serviceEntryLifecycleDelegate) Updated(obj *ServiceEntry) (runtime.Object, error) {
	if n.update == nil {
		return obj, nil
	}
	return n.update(obj)
}
