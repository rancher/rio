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
	ServiceScaleRecommendationGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "ServiceScaleRecommendation",
	}
	ServiceScaleRecommendationResource = metav1.APIResource{
		Name:         "servicescalerecommendations",
		SingularName: "servicescalerecommendation",
		Namespaced:   true,

		Kind: ServiceScaleRecommendationGroupVersionKind.Kind,
	}
)

func NewServiceScaleRecommendation(namespace, name string, obj ServiceScaleRecommendation) *ServiceScaleRecommendation {
	obj.APIVersion, obj.Kind = ServiceScaleRecommendationGroupVersionKind.ToAPIVersionAndKind()
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

type ServiceScaleRecommendationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceScaleRecommendation
}

type ServiceScaleRecommendationHandlerFunc func(key string, obj *ServiceScaleRecommendation) (runtime.Object, error)

type ServiceScaleRecommendationChangeHandlerFunc func(obj *ServiceScaleRecommendation) (runtime.Object, error)

type ServiceScaleRecommendationLister interface {
	List(namespace string, selector labels.Selector) (ret []*ServiceScaleRecommendation, err error)
	Get(namespace, name string) (*ServiceScaleRecommendation, error)
}

type ServiceScaleRecommendationController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() ServiceScaleRecommendationLister
	AddHandler(ctx context.Context, name string, handler ServiceScaleRecommendationHandlerFunc)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, handler ServiceScaleRecommendationHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type ServiceScaleRecommendationInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*ServiceScaleRecommendation) (*ServiceScaleRecommendation, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*ServiceScaleRecommendation, error)
	Get(name string, opts metav1.GetOptions) (*ServiceScaleRecommendation, error)
	Update(*ServiceScaleRecommendation) (*ServiceScaleRecommendation, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*ServiceScaleRecommendationList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() ServiceScaleRecommendationController
	AddHandler(ctx context.Context, name string, sync ServiceScaleRecommendationHandlerFunc)
	AddLifecycle(ctx context.Context, name string, lifecycle ServiceScaleRecommendationLifecycle)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync ServiceScaleRecommendationHandlerFunc)
	AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle ServiceScaleRecommendationLifecycle)
}

type serviceScaleRecommendationLister struct {
	controller *serviceScaleRecommendationController
}

func (l *serviceScaleRecommendationLister) List(namespace string, selector labels.Selector) (ret []*ServiceScaleRecommendation, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*ServiceScaleRecommendation))
	})
	return
}

func (l *serviceScaleRecommendationLister) Get(namespace, name string) (*ServiceScaleRecommendation, error) {
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
			Group:    ServiceScaleRecommendationGroupVersionKind.Group,
			Resource: "serviceScaleRecommendation",
		}, key)
	}
	return obj.(*ServiceScaleRecommendation), nil
}

type serviceScaleRecommendationController struct {
	controller.GenericController
}

func (c *serviceScaleRecommendationController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *serviceScaleRecommendationController) Lister() ServiceScaleRecommendationLister {
	return &serviceScaleRecommendationLister{
		controller: c,
	}
}

func (c *serviceScaleRecommendationController) AddHandler(ctx context.Context, name string, handler ServiceScaleRecommendationHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*ServiceScaleRecommendation); ok {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

func (c *serviceScaleRecommendationController) AddClusterScopedHandler(ctx context.Context, name, cluster string, handler ServiceScaleRecommendationHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*ServiceScaleRecommendation); ok && controller.ObjectInCluster(cluster, obj) {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

type serviceScaleRecommendationFactory struct {
}

func (c serviceScaleRecommendationFactory) Object() runtime.Object {
	return &ServiceScaleRecommendation{}
}

func (c serviceScaleRecommendationFactory) List() runtime.Object {
	return &ServiceScaleRecommendationList{}
}

func (s *serviceScaleRecommendationClient) Controller() ServiceScaleRecommendationController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.serviceScaleRecommendationControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(ServiceScaleRecommendationGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &serviceScaleRecommendationController{
		GenericController: genericController,
	}

	s.client.serviceScaleRecommendationControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type serviceScaleRecommendationClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   ServiceScaleRecommendationController
}

func (s *serviceScaleRecommendationClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *serviceScaleRecommendationClient) Create(o *ServiceScaleRecommendation) (*ServiceScaleRecommendation, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*ServiceScaleRecommendation), err
}

func (s *serviceScaleRecommendationClient) Get(name string, opts metav1.GetOptions) (*ServiceScaleRecommendation, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*ServiceScaleRecommendation), err
}

func (s *serviceScaleRecommendationClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*ServiceScaleRecommendation, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*ServiceScaleRecommendation), err
}

func (s *serviceScaleRecommendationClient) Update(o *ServiceScaleRecommendation) (*ServiceScaleRecommendation, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*ServiceScaleRecommendation), err
}

func (s *serviceScaleRecommendationClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *serviceScaleRecommendationClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *serviceScaleRecommendationClient) List(opts metav1.ListOptions) (*ServiceScaleRecommendationList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*ServiceScaleRecommendationList), err
}

func (s *serviceScaleRecommendationClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *serviceScaleRecommendationClient) Patch(o *ServiceScaleRecommendation, patchType types.PatchType, data []byte, subresources ...string) (*ServiceScaleRecommendation, error) {
	obj, err := s.objectClient.Patch(o.Name, o, patchType, data, subresources...)
	return obj.(*ServiceScaleRecommendation), err
}

func (s *serviceScaleRecommendationClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *serviceScaleRecommendationClient) AddHandler(ctx context.Context, name string, sync ServiceScaleRecommendationHandlerFunc) {
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *serviceScaleRecommendationClient) AddLifecycle(ctx context.Context, name string, lifecycle ServiceScaleRecommendationLifecycle) {
	sync := NewServiceScaleRecommendationLifecycleAdapter(name, false, s, lifecycle)
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *serviceScaleRecommendationClient) AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync ServiceScaleRecommendationHandlerFunc) {
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

func (s *serviceScaleRecommendationClient) AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle ServiceScaleRecommendationLifecycle) {
	sync := NewServiceScaleRecommendationLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

type ServiceScaleRecommendationIndexer func(obj *ServiceScaleRecommendation) ([]string, error)

type ServiceScaleRecommendationClientCache interface {
	Get(namespace, name string) (*ServiceScaleRecommendation, error)
	List(namespace string, selector labels.Selector) ([]*ServiceScaleRecommendation, error)

	Index(name string, indexer ServiceScaleRecommendationIndexer)
	GetIndexed(name, key string) ([]*ServiceScaleRecommendation, error)
}

type ServiceScaleRecommendationClient interface {
	Create(*ServiceScaleRecommendation) (*ServiceScaleRecommendation, error)
	Get(namespace, name string, opts metav1.GetOptions) (*ServiceScaleRecommendation, error)
	Update(*ServiceScaleRecommendation) (*ServiceScaleRecommendation, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	List(namespace string, opts metav1.ListOptions) (*ServiceScaleRecommendationList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)

	Cache() ServiceScaleRecommendationClientCache

	OnCreate(ctx context.Context, name string, sync ServiceScaleRecommendationChangeHandlerFunc)
	OnChange(ctx context.Context, name string, sync ServiceScaleRecommendationChangeHandlerFunc)
	OnRemove(ctx context.Context, name string, sync ServiceScaleRecommendationChangeHandlerFunc)
	Enqueue(namespace, name string)

	Generic() controller.GenericController
	ObjectClient() *objectclient.ObjectClient
	Interface() ServiceScaleRecommendationInterface
}

type serviceScaleRecommendationClientCache struct {
	client *serviceScaleRecommendationClient2
}

type serviceScaleRecommendationClient2 struct {
	iface      ServiceScaleRecommendationInterface
	controller ServiceScaleRecommendationController
}

func (n *serviceScaleRecommendationClient2) Interface() ServiceScaleRecommendationInterface {
	return n.iface
}

func (n *serviceScaleRecommendationClient2) Generic() controller.GenericController {
	return n.iface.Controller().Generic()
}

func (n *serviceScaleRecommendationClient2) ObjectClient() *objectclient.ObjectClient {
	return n.Interface().ObjectClient()
}

func (n *serviceScaleRecommendationClient2) Enqueue(namespace, name string) {
	n.iface.Controller().Enqueue(namespace, name)
}

func (n *serviceScaleRecommendationClient2) Create(obj *ServiceScaleRecommendation) (*ServiceScaleRecommendation, error) {
	return n.iface.Create(obj)
}

func (n *serviceScaleRecommendationClient2) Get(namespace, name string, opts metav1.GetOptions) (*ServiceScaleRecommendation, error) {
	return n.iface.GetNamespaced(namespace, name, opts)
}

func (n *serviceScaleRecommendationClient2) Update(obj *ServiceScaleRecommendation) (*ServiceScaleRecommendation, error) {
	return n.iface.Update(obj)
}

func (n *serviceScaleRecommendationClient2) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	return n.iface.DeleteNamespaced(namespace, name, options)
}

func (n *serviceScaleRecommendationClient2) List(namespace string, opts metav1.ListOptions) (*ServiceScaleRecommendationList, error) {
	return n.iface.List(opts)
}

func (n *serviceScaleRecommendationClient2) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return n.iface.Watch(opts)
}

func (n *serviceScaleRecommendationClientCache) Get(namespace, name string) (*ServiceScaleRecommendation, error) {
	return n.client.controller.Lister().Get(namespace, name)
}

func (n *serviceScaleRecommendationClientCache) List(namespace string, selector labels.Selector) ([]*ServiceScaleRecommendation, error) {
	return n.client.controller.Lister().List(namespace, selector)
}

func (n *serviceScaleRecommendationClient2) Cache() ServiceScaleRecommendationClientCache {
	n.loadController()
	return &serviceScaleRecommendationClientCache{
		client: n,
	}
}

func (n *serviceScaleRecommendationClient2) OnCreate(ctx context.Context, name string, sync ServiceScaleRecommendationChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-create", &serviceScaleRecommendationLifecycleDelegate{create: sync})
}

func (n *serviceScaleRecommendationClient2) OnChange(ctx context.Context, name string, sync ServiceScaleRecommendationChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-change", &serviceScaleRecommendationLifecycleDelegate{update: sync})
}

func (n *serviceScaleRecommendationClient2) OnRemove(ctx context.Context, name string, sync ServiceScaleRecommendationChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name, &serviceScaleRecommendationLifecycleDelegate{remove: sync})
}

func (n *serviceScaleRecommendationClientCache) Index(name string, indexer ServiceScaleRecommendationIndexer) {
	err := n.client.controller.Informer().GetIndexer().AddIndexers(map[string]cache.IndexFunc{
		name: func(obj interface{}) ([]string, error) {
			if v, ok := obj.(*ServiceScaleRecommendation); ok {
				return indexer(v)
			}
			return nil, nil
		},
	})

	if err != nil {
		panic(err)
	}
}

func (n *serviceScaleRecommendationClientCache) GetIndexed(name, key string) ([]*ServiceScaleRecommendation, error) {
	var result []*ServiceScaleRecommendation
	objs, err := n.client.controller.Informer().GetIndexer().ByIndex(name, key)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		if v, ok := obj.(*ServiceScaleRecommendation); ok {
			result = append(result, v)
		}
	}

	return result, nil
}

func (n *serviceScaleRecommendationClient2) loadController() {
	if n.controller == nil {
		n.controller = n.iface.Controller()
	}
}

type serviceScaleRecommendationLifecycleDelegate struct {
	create ServiceScaleRecommendationChangeHandlerFunc
	update ServiceScaleRecommendationChangeHandlerFunc
	remove ServiceScaleRecommendationChangeHandlerFunc
}

func (n *serviceScaleRecommendationLifecycleDelegate) HasCreate() bool {
	return n.create != nil
}

func (n *serviceScaleRecommendationLifecycleDelegate) Create(obj *ServiceScaleRecommendation) (runtime.Object, error) {
	if n.create == nil {
		return obj, nil
	}
	return n.create(obj)
}

func (n *serviceScaleRecommendationLifecycleDelegate) HasFinalize() bool {
	return n.remove != nil
}

func (n *serviceScaleRecommendationLifecycleDelegate) Remove(obj *ServiceScaleRecommendation) (runtime.Object, error) {
	if n.remove == nil {
		return obj, nil
	}
	return n.remove(obj)
}

func (n *serviceScaleRecommendationLifecycleDelegate) Updated(obj *ServiceScaleRecommendation) (runtime.Object, error) {
	if n.update == nil {
		return obj, nil
	}
	return n.update(obj)
}
