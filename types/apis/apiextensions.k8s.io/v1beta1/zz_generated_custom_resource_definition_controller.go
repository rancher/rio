package v1beta1

import (
	"context"

	"github.com/rancher/norman/controller"
	"github.com/rancher/norman/objectclient"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
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
	CustomResourceDefinitionGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "CustomResourceDefinition",
	}
	CustomResourceDefinitionResource = metav1.APIResource{
		Name:         "customresourcedefinitions",
		SingularName: "customresourcedefinition",
		Namespaced:   false,
		Kind:         CustomResourceDefinitionGroupVersionKind.Kind,
	}
)

func NewCustomResourceDefinition(namespace, name string, obj v1beta1.CustomResourceDefinition) *v1beta1.CustomResourceDefinition {
	obj.APIVersion, obj.Kind = CustomResourceDefinitionGroupVersionKind.ToAPIVersionAndKind()
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

type CustomResourceDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []v1beta1.CustomResourceDefinition
}

type CustomResourceDefinitionHandlerFunc func(key string, obj *v1beta1.CustomResourceDefinition) (runtime.Object, error)

type CustomResourceDefinitionChangeHandlerFunc func(obj *v1beta1.CustomResourceDefinition) (runtime.Object, error)

type CustomResourceDefinitionLister interface {
	List(namespace string, selector labels.Selector) (ret []*v1beta1.CustomResourceDefinition, err error)
	Get(namespace, name string) (*v1beta1.CustomResourceDefinition, error)
}

type CustomResourceDefinitionController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() CustomResourceDefinitionLister
	AddHandler(ctx context.Context, name string, handler CustomResourceDefinitionHandlerFunc)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, handler CustomResourceDefinitionHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type CustomResourceDefinitionInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*v1beta1.CustomResourceDefinition) (*v1beta1.CustomResourceDefinition, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*v1beta1.CustomResourceDefinition, error)
	Get(name string, opts metav1.GetOptions) (*v1beta1.CustomResourceDefinition, error)
	Update(*v1beta1.CustomResourceDefinition) (*v1beta1.CustomResourceDefinition, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*CustomResourceDefinitionList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() CustomResourceDefinitionController
	AddHandler(ctx context.Context, name string, sync CustomResourceDefinitionHandlerFunc)
	AddLifecycle(ctx context.Context, name string, lifecycle CustomResourceDefinitionLifecycle)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync CustomResourceDefinitionHandlerFunc)
	AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle CustomResourceDefinitionLifecycle)
}

type customResourceDefinitionLister struct {
	controller *customResourceDefinitionController
}

func (l *customResourceDefinitionLister) List(namespace string, selector labels.Selector) (ret []*v1beta1.CustomResourceDefinition, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*v1beta1.CustomResourceDefinition))
	})
	return
}

func (l *customResourceDefinitionLister) Get(namespace, name string) (*v1beta1.CustomResourceDefinition, error) {
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
			Group:    CustomResourceDefinitionGroupVersionKind.Group,
			Resource: "customResourceDefinition",
		}, key)
	}
	return obj.(*v1beta1.CustomResourceDefinition), nil
}

type customResourceDefinitionController struct {
	controller.GenericController
}

func (c *customResourceDefinitionController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *customResourceDefinitionController) Lister() CustomResourceDefinitionLister {
	return &customResourceDefinitionLister{
		controller: c,
	}
}

func (c *customResourceDefinitionController) AddHandler(ctx context.Context, name string, handler CustomResourceDefinitionHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*v1beta1.CustomResourceDefinition); ok {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

func (c *customResourceDefinitionController) AddClusterScopedHandler(ctx context.Context, name, cluster string, handler CustomResourceDefinitionHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*v1beta1.CustomResourceDefinition); ok && controller.ObjectInCluster(cluster, obj) {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

type customResourceDefinitionFactory struct {
}

func (c customResourceDefinitionFactory) Object() runtime.Object {
	return &v1beta1.CustomResourceDefinition{}
}

func (c customResourceDefinitionFactory) List() runtime.Object {
	return &CustomResourceDefinitionList{}
}

func (s *customResourceDefinitionClient) Controller() CustomResourceDefinitionController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.customResourceDefinitionControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(CustomResourceDefinitionGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &customResourceDefinitionController{
		GenericController: genericController,
	}

	s.client.customResourceDefinitionControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type customResourceDefinitionClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   CustomResourceDefinitionController
}

func (s *customResourceDefinitionClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *customResourceDefinitionClient) Create(o *v1beta1.CustomResourceDefinition) (*v1beta1.CustomResourceDefinition, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*v1beta1.CustomResourceDefinition), err
}

func (s *customResourceDefinitionClient) Get(name string, opts metav1.GetOptions) (*v1beta1.CustomResourceDefinition, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*v1beta1.CustomResourceDefinition), err
}

func (s *customResourceDefinitionClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*v1beta1.CustomResourceDefinition, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*v1beta1.CustomResourceDefinition), err
}

func (s *customResourceDefinitionClient) Update(o *v1beta1.CustomResourceDefinition) (*v1beta1.CustomResourceDefinition, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*v1beta1.CustomResourceDefinition), err
}

func (s *customResourceDefinitionClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *customResourceDefinitionClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *customResourceDefinitionClient) List(opts metav1.ListOptions) (*CustomResourceDefinitionList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*CustomResourceDefinitionList), err
}

func (s *customResourceDefinitionClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *customResourceDefinitionClient) Patch(o *v1beta1.CustomResourceDefinition, patchType types.PatchType, data []byte, subresources ...string) (*v1beta1.CustomResourceDefinition, error) {
	obj, err := s.objectClient.Patch(o.Name, o, patchType, data, subresources...)
	return obj.(*v1beta1.CustomResourceDefinition), err
}

func (s *customResourceDefinitionClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *customResourceDefinitionClient) AddHandler(ctx context.Context, name string, sync CustomResourceDefinitionHandlerFunc) {
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *customResourceDefinitionClient) AddLifecycle(ctx context.Context, name string, lifecycle CustomResourceDefinitionLifecycle) {
	sync := NewCustomResourceDefinitionLifecycleAdapter(name, false, s, lifecycle)
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *customResourceDefinitionClient) AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync CustomResourceDefinitionHandlerFunc) {
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

func (s *customResourceDefinitionClient) AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle CustomResourceDefinitionLifecycle) {
	sync := NewCustomResourceDefinitionLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

type CustomResourceDefinitionIndexer func(obj *v1beta1.CustomResourceDefinition) ([]string, error)

type CustomResourceDefinitionClientCache interface {
	Get(namespace, name string) (*v1beta1.CustomResourceDefinition, error)
	List(namespace string, selector labels.Selector) ([]*v1beta1.CustomResourceDefinition, error)

	Index(name string, indexer CustomResourceDefinitionIndexer)
	GetIndexed(name, key string) ([]*v1beta1.CustomResourceDefinition, error)
}

type CustomResourceDefinitionClient interface {
	Create(*v1beta1.CustomResourceDefinition) (*v1beta1.CustomResourceDefinition, error)
	Get(namespace, name string, opts metav1.GetOptions) (*v1beta1.CustomResourceDefinition, error)
	Update(*v1beta1.CustomResourceDefinition) (*v1beta1.CustomResourceDefinition, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	List(namespace string, opts metav1.ListOptions) (*CustomResourceDefinitionList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)

	Cache() CustomResourceDefinitionClientCache

	OnCreate(ctx context.Context, name string, sync CustomResourceDefinitionChangeHandlerFunc)
	OnChange(ctx context.Context, name string, sync CustomResourceDefinitionChangeHandlerFunc)
	OnRemove(ctx context.Context, name string, sync CustomResourceDefinitionChangeHandlerFunc)
	Enqueue(namespace, name string)

	Generic() controller.GenericController
	ObjectClient() *objectclient.ObjectClient
	Interface() CustomResourceDefinitionInterface
}

type customResourceDefinitionClientCache struct {
	client *customResourceDefinitionClient2
}

type customResourceDefinitionClient2 struct {
	iface      CustomResourceDefinitionInterface
	controller CustomResourceDefinitionController
}

func (n *customResourceDefinitionClient2) Interface() CustomResourceDefinitionInterface {
	return n.iface
}

func (n *customResourceDefinitionClient2) Generic() controller.GenericController {
	return n.iface.Controller().Generic()
}

func (n *customResourceDefinitionClient2) ObjectClient() *objectclient.ObjectClient {
	return n.Interface().ObjectClient()
}

func (n *customResourceDefinitionClient2) Enqueue(namespace, name string) {
	n.iface.Controller().Enqueue(namespace, name)
}

func (n *customResourceDefinitionClient2) Create(obj *v1beta1.CustomResourceDefinition) (*v1beta1.CustomResourceDefinition, error) {
	return n.iface.Create(obj)
}

func (n *customResourceDefinitionClient2) Get(namespace, name string, opts metav1.GetOptions) (*v1beta1.CustomResourceDefinition, error) {
	return n.iface.GetNamespaced(namespace, name, opts)
}

func (n *customResourceDefinitionClient2) Update(obj *v1beta1.CustomResourceDefinition) (*v1beta1.CustomResourceDefinition, error) {
	return n.iface.Update(obj)
}

func (n *customResourceDefinitionClient2) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	return n.iface.DeleteNamespaced(namespace, name, options)
}

func (n *customResourceDefinitionClient2) List(namespace string, opts metav1.ListOptions) (*CustomResourceDefinitionList, error) {
	return n.iface.List(opts)
}

func (n *customResourceDefinitionClient2) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return n.iface.Watch(opts)
}

func (n *customResourceDefinitionClientCache) Get(namespace, name string) (*v1beta1.CustomResourceDefinition, error) {
	return n.client.controller.Lister().Get(namespace, name)
}

func (n *customResourceDefinitionClientCache) List(namespace string, selector labels.Selector) ([]*v1beta1.CustomResourceDefinition, error) {
	return n.client.controller.Lister().List(namespace, selector)
}

func (n *customResourceDefinitionClient2) Cache() CustomResourceDefinitionClientCache {
	n.loadController()
	return &customResourceDefinitionClientCache{
		client: n,
	}
}

func (n *customResourceDefinitionClient2) OnCreate(ctx context.Context, name string, sync CustomResourceDefinitionChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-create", &customResourceDefinitionLifecycleDelegate{create: sync})
}

func (n *customResourceDefinitionClient2) OnChange(ctx context.Context, name string, sync CustomResourceDefinitionChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-change", &customResourceDefinitionLifecycleDelegate{update: sync})
}

func (n *customResourceDefinitionClient2) OnRemove(ctx context.Context, name string, sync CustomResourceDefinitionChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name, &customResourceDefinitionLifecycleDelegate{remove: sync})
}

func (n *customResourceDefinitionClientCache) Index(name string, indexer CustomResourceDefinitionIndexer) {
	err := n.client.controller.Informer().GetIndexer().AddIndexers(map[string]cache.IndexFunc{
		name: func(obj interface{}) ([]string, error) {
			if v, ok := obj.(*v1beta1.CustomResourceDefinition); ok {
				return indexer(v)
			}
			return nil, nil
		},
	})

	if err != nil {
		panic(err)
	}
}

func (n *customResourceDefinitionClientCache) GetIndexed(name, key string) ([]*v1beta1.CustomResourceDefinition, error) {
	var result []*v1beta1.CustomResourceDefinition
	objs, err := n.client.controller.Informer().GetIndexer().ByIndex(name, key)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		if v, ok := obj.(*v1beta1.CustomResourceDefinition); ok {
			result = append(result, v)
		}
	}

	return result, nil
}

func (n *customResourceDefinitionClient2) loadController() {
	if n.controller == nil {
		n.controller = n.iface.Controller()
	}
}

type customResourceDefinitionLifecycleDelegate struct {
	create CustomResourceDefinitionChangeHandlerFunc
	update CustomResourceDefinitionChangeHandlerFunc
	remove CustomResourceDefinitionChangeHandlerFunc
}

func (n *customResourceDefinitionLifecycleDelegate) HasCreate() bool {
	return n.create != nil
}

func (n *customResourceDefinitionLifecycleDelegate) Create(obj *v1beta1.CustomResourceDefinition) (runtime.Object, error) {
	if n.create == nil {
		return obj, nil
	}
	return n.create(obj)
}

func (n *customResourceDefinitionLifecycleDelegate) HasFinalize() bool {
	return n.remove != nil
}

func (n *customResourceDefinitionLifecycleDelegate) Remove(obj *v1beta1.CustomResourceDefinition) (runtime.Object, error) {
	if n.remove == nil {
		return obj, nil
	}
	return n.remove(obj)
}

func (n *customResourceDefinitionLifecycleDelegate) Updated(obj *v1beta1.CustomResourceDefinition) (runtime.Object, error) {
	if n.update == nil {
		return obj, nil
	}
	return n.update(obj)
}
