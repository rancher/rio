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
	ExternalServiceGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "ExternalService",
	}
	ExternalServiceResource = metav1.APIResource{
		Name:         "externalservices",
		SingularName: "externalservice",
		Namespaced:   true,

		Kind: ExternalServiceGroupVersionKind.Kind,
	}
)

func NewExternalService(namespace, name string, obj ExternalService) *ExternalService {
	obj.APIVersion, obj.Kind = ExternalServiceGroupVersionKind.ToAPIVersionAndKind()
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

type ExternalServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExternalService
}

type ExternalServiceHandlerFunc func(key string, obj *ExternalService) (runtime.Object, error)

type ExternalServiceChangeHandlerFunc func(obj *ExternalService) (runtime.Object, error)

type ExternalServiceLister interface {
	List(namespace string, selector labels.Selector) (ret []*ExternalService, err error)
	Get(namespace, name string) (*ExternalService, error)
}

type ExternalServiceController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() ExternalServiceLister
	AddHandler(ctx context.Context, name string, handler ExternalServiceHandlerFunc)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, handler ExternalServiceHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type ExternalServiceInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*ExternalService) (*ExternalService, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*ExternalService, error)
	Get(name string, opts metav1.GetOptions) (*ExternalService, error)
	Update(*ExternalService) (*ExternalService, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*ExternalServiceList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() ExternalServiceController
	AddHandler(ctx context.Context, name string, sync ExternalServiceHandlerFunc)
	AddLifecycle(ctx context.Context, name string, lifecycle ExternalServiceLifecycle)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync ExternalServiceHandlerFunc)
	AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle ExternalServiceLifecycle)
}

type externalServiceLister struct {
	controller *externalServiceController
}

func (l *externalServiceLister) List(namespace string, selector labels.Selector) (ret []*ExternalService, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*ExternalService))
	})
	return
}

func (l *externalServiceLister) Get(namespace, name string) (*ExternalService, error) {
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
			Group:    ExternalServiceGroupVersionKind.Group,
			Resource: "externalService",
		}, key)
	}
	return obj.(*ExternalService), nil
}

type externalServiceController struct {
	controller.GenericController
}

func (c *externalServiceController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *externalServiceController) Lister() ExternalServiceLister {
	return &externalServiceLister{
		controller: c,
	}
}

func (c *externalServiceController) AddHandler(ctx context.Context, name string, handler ExternalServiceHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*ExternalService); ok {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

func (c *externalServiceController) AddClusterScopedHandler(ctx context.Context, name, cluster string, handler ExternalServiceHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*ExternalService); ok && controller.ObjectInCluster(cluster, obj) {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

type externalServiceFactory struct {
}

func (c externalServiceFactory) Object() runtime.Object {
	return &ExternalService{}
}

func (c externalServiceFactory) List() runtime.Object {
	return &ExternalServiceList{}
}

func (s *externalServiceClient) Controller() ExternalServiceController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.externalServiceControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(ExternalServiceGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &externalServiceController{
		GenericController: genericController,
	}

	s.client.externalServiceControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type externalServiceClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   ExternalServiceController
}

func (s *externalServiceClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *externalServiceClient) Create(o *ExternalService) (*ExternalService, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*ExternalService), err
}

func (s *externalServiceClient) Get(name string, opts metav1.GetOptions) (*ExternalService, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*ExternalService), err
}

func (s *externalServiceClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*ExternalService, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*ExternalService), err
}

func (s *externalServiceClient) Update(o *ExternalService) (*ExternalService, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*ExternalService), err
}

func (s *externalServiceClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *externalServiceClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *externalServiceClient) List(opts metav1.ListOptions) (*ExternalServiceList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*ExternalServiceList), err
}

func (s *externalServiceClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *externalServiceClient) Patch(o *ExternalService, patchType types.PatchType, data []byte, subresources ...string) (*ExternalService, error) {
	obj, err := s.objectClient.Patch(o.Name, o, patchType, data, subresources...)
	return obj.(*ExternalService), err
}

func (s *externalServiceClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *externalServiceClient) AddHandler(ctx context.Context, name string, sync ExternalServiceHandlerFunc) {
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *externalServiceClient) AddLifecycle(ctx context.Context, name string, lifecycle ExternalServiceLifecycle) {
	sync := NewExternalServiceLifecycleAdapter(name, false, s, lifecycle)
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *externalServiceClient) AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync ExternalServiceHandlerFunc) {
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

func (s *externalServiceClient) AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle ExternalServiceLifecycle) {
	sync := NewExternalServiceLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

type ExternalServiceIndexer func(obj *ExternalService) ([]string, error)

type ExternalServiceClientCache interface {
	Get(namespace, name string) (*ExternalService, error)
	List(namespace string, selector labels.Selector) ([]*ExternalService, error)

	Index(name string, indexer ExternalServiceIndexer)
	GetIndexed(name, key string) ([]*ExternalService, error)
}

type ExternalServiceClient interface {
	Create(*ExternalService) (*ExternalService, error)
	Get(namespace, name string, opts metav1.GetOptions) (*ExternalService, error)
	Update(*ExternalService) (*ExternalService, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	List(namespace string, opts metav1.ListOptions) (*ExternalServiceList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)

	Cache() ExternalServiceClientCache

	OnCreate(ctx context.Context, name string, sync ExternalServiceChangeHandlerFunc)
	OnChange(ctx context.Context, name string, sync ExternalServiceChangeHandlerFunc)
	OnRemove(ctx context.Context, name string, sync ExternalServiceChangeHandlerFunc)
	Enqueue(namespace, name string)

	Generic() controller.GenericController
	ObjectClient() *objectclient.ObjectClient
	Interface() ExternalServiceInterface
}

type externalServiceClientCache struct {
	client *externalServiceClient2
}

type externalServiceClient2 struct {
	iface      ExternalServiceInterface
	controller ExternalServiceController
}

func (n *externalServiceClient2) Interface() ExternalServiceInterface {
	return n.iface
}

func (n *externalServiceClient2) Generic() controller.GenericController {
	return n.iface.Controller().Generic()
}

func (n *externalServiceClient2) ObjectClient() *objectclient.ObjectClient {
	return n.Interface().ObjectClient()
}

func (n *externalServiceClient2) Enqueue(namespace, name string) {
	n.iface.Controller().Enqueue(namespace, name)
}

func (n *externalServiceClient2) Create(obj *ExternalService) (*ExternalService, error) {
	return n.iface.Create(obj)
}

func (n *externalServiceClient2) Get(namespace, name string, opts metav1.GetOptions) (*ExternalService, error) {
	return n.iface.GetNamespaced(namespace, name, opts)
}

func (n *externalServiceClient2) Update(obj *ExternalService) (*ExternalService, error) {
	return n.iface.Update(obj)
}

func (n *externalServiceClient2) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	return n.iface.DeleteNamespaced(namespace, name, options)
}

func (n *externalServiceClient2) List(namespace string, opts metav1.ListOptions) (*ExternalServiceList, error) {
	return n.iface.List(opts)
}

func (n *externalServiceClient2) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return n.iface.Watch(opts)
}

func (n *externalServiceClientCache) Get(namespace, name string) (*ExternalService, error) {
	return n.client.controller.Lister().Get(namespace, name)
}

func (n *externalServiceClientCache) List(namespace string, selector labels.Selector) ([]*ExternalService, error) {
	return n.client.controller.Lister().List(namespace, selector)
}

func (n *externalServiceClient2) Cache() ExternalServiceClientCache {
	n.loadController()
	return &externalServiceClientCache{
		client: n,
	}
}

func (n *externalServiceClient2) OnCreate(ctx context.Context, name string, sync ExternalServiceChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-create", &externalServiceLifecycleDelegate{create: sync})
}

func (n *externalServiceClient2) OnChange(ctx context.Context, name string, sync ExternalServiceChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-change", &externalServiceLifecycleDelegate{update: sync})
}

func (n *externalServiceClient2) OnRemove(ctx context.Context, name string, sync ExternalServiceChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name, &externalServiceLifecycleDelegate{remove: sync})
}

func (n *externalServiceClientCache) Index(name string, indexer ExternalServiceIndexer) {
	err := n.client.controller.Informer().GetIndexer().AddIndexers(map[string]cache.IndexFunc{
		name: func(obj interface{}) ([]string, error) {
			if v, ok := obj.(*ExternalService); ok {
				return indexer(v)
			}
			return nil, nil
		},
	})

	if err != nil {
		panic(err)
	}
}

func (n *externalServiceClientCache) GetIndexed(name, key string) ([]*ExternalService, error) {
	var result []*ExternalService
	objs, err := n.client.controller.Informer().GetIndexer().ByIndex(name, key)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		if v, ok := obj.(*ExternalService); ok {
			result = append(result, v)
		}
	}

	return result, nil
}

func (n *externalServiceClient2) loadController() {
	if n.controller == nil {
		n.controller = n.iface.Controller()
	}
}

type externalServiceLifecycleDelegate struct {
	create ExternalServiceChangeHandlerFunc
	update ExternalServiceChangeHandlerFunc
	remove ExternalServiceChangeHandlerFunc
}

func (n *externalServiceLifecycleDelegate) HasCreate() bool {
	return n.create != nil
}

func (n *externalServiceLifecycleDelegate) Create(obj *ExternalService) (runtime.Object, error) {
	if n.create == nil {
		return obj, nil
	}
	return n.create(obj)
}

func (n *externalServiceLifecycleDelegate) HasFinalize() bool {
	return n.remove != nil
}

func (n *externalServiceLifecycleDelegate) Remove(obj *ExternalService) (runtime.Object, error) {
	if n.remove == nil {
		return obj, nil
	}
	return n.remove(obj)
}

func (n *externalServiceLifecycleDelegate) Updated(obj *ExternalService) (runtime.Object, error) {
	if n.update == nil {
		return obj, nil
	}
	return n.update(obj)
}
