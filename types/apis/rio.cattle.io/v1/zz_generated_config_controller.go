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
	ConfigGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "Config",
	}
	ConfigResource = metav1.APIResource{
		Name:         "configs",
		SingularName: "config",
		Namespaced:   true,

		Kind: ConfigGroupVersionKind.Kind,
	}
)

func NewConfig(namespace, name string, obj Config) *Config {
	obj.APIVersion, obj.Kind = ConfigGroupVersionKind.ToAPIVersionAndKind()
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

type ConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Config
}

type ConfigHandlerFunc func(key string, obj *Config) (runtime.Object, error)

type ConfigChangeHandlerFunc func(obj *Config) (runtime.Object, error)

type ConfigLister interface {
	List(namespace string, selector labels.Selector) (ret []*Config, err error)
	Get(namespace, name string) (*Config, error)
}

type ConfigController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() ConfigLister
	AddHandler(ctx context.Context, name string, handler ConfigHandlerFunc)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, handler ConfigHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type ConfigInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*Config) (*Config, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*Config, error)
	Get(name string, opts metav1.GetOptions) (*Config, error)
	Update(*Config) (*Config, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*ConfigList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() ConfigController
	AddHandler(ctx context.Context, name string, sync ConfigHandlerFunc)
	AddLifecycle(ctx context.Context, name string, lifecycle ConfigLifecycle)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync ConfigHandlerFunc)
	AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle ConfigLifecycle)
}

type configLister struct {
	controller *configController
}

func (l *configLister) List(namespace string, selector labels.Selector) (ret []*Config, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*Config))
	})
	return
}

func (l *configLister) Get(namespace, name string) (*Config, error) {
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
			Group:    ConfigGroupVersionKind.Group,
			Resource: "config",
		}, key)
	}
	return obj.(*Config), nil
}

type configController struct {
	controller.GenericController
}

func (c *configController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *configController) Lister() ConfigLister {
	return &configLister{
		controller: c,
	}
}

func (c *configController) AddHandler(ctx context.Context, name string, handler ConfigHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*Config); ok {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

func (c *configController) AddClusterScopedHandler(ctx context.Context, name, cluster string, handler ConfigHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*Config); ok && controller.ObjectInCluster(cluster, obj) {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

type configFactory struct {
}

func (c configFactory) Object() runtime.Object {
	return &Config{}
}

func (c configFactory) List() runtime.Object {
	return &ConfigList{}
}

func (s *configClient) Controller() ConfigController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.configControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(ConfigGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &configController{
		GenericController: genericController,
	}

	s.client.configControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type configClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   ConfigController
}

func (s *configClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *configClient) Create(o *Config) (*Config, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*Config), err
}

func (s *configClient) Get(name string, opts metav1.GetOptions) (*Config, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*Config), err
}

func (s *configClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*Config, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*Config), err
}

func (s *configClient) Update(o *Config) (*Config, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*Config), err
}

func (s *configClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *configClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *configClient) List(opts metav1.ListOptions) (*ConfigList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*ConfigList), err
}

func (s *configClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *configClient) Patch(o *Config, patchType types.PatchType, data []byte, subresources ...string) (*Config, error) {
	obj, err := s.objectClient.Patch(o.Name, o, patchType, data, subresources...)
	return obj.(*Config), err
}

func (s *configClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *configClient) AddHandler(ctx context.Context, name string, sync ConfigHandlerFunc) {
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *configClient) AddLifecycle(ctx context.Context, name string, lifecycle ConfigLifecycle) {
	sync := NewConfigLifecycleAdapter(name, false, s, lifecycle)
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *configClient) AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync ConfigHandlerFunc) {
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

func (s *configClient) AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle ConfigLifecycle) {
	sync := NewConfigLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

type ConfigIndexer func(obj *Config) ([]string, error)

type ConfigClientCache interface {
	Get(namespace, name string) (*Config, error)
	List(namespace string, selector labels.Selector) ([]*Config, error)

	Index(name string, indexer ConfigIndexer)
	GetIndexed(name, key string) ([]*Config, error)
}

type ConfigClient interface {
	Create(*Config) (*Config, error)
	Get(namespace, name string, opts metav1.GetOptions) (*Config, error)
	Update(*Config) (*Config, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	List(namespace string, opts metav1.ListOptions) (*ConfigList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)

	Cache() ConfigClientCache

	OnCreate(ctx context.Context, name string, sync ConfigChangeHandlerFunc)
	OnChange(ctx context.Context, name string, sync ConfigChangeHandlerFunc)
	OnRemove(ctx context.Context, name string, sync ConfigChangeHandlerFunc)
	Enqueue(namespace, name string)

	Generic() controller.GenericController
	ObjectClient() *objectclient.ObjectClient
	Interface() ConfigInterface
}

type configClientCache struct {
	client *configClient2
}

type configClient2 struct {
	iface      ConfigInterface
	controller ConfigController
}

func (n *configClient2) Interface() ConfigInterface {
	return n.iface
}

func (n *configClient2) Generic() controller.GenericController {
	return n.iface.Controller().Generic()
}

func (n *configClient2) ObjectClient() *objectclient.ObjectClient {
	return n.Interface().ObjectClient()
}

func (n *configClient2) Enqueue(namespace, name string) {
	n.iface.Controller().Enqueue(namespace, name)
}

func (n *configClient2) Create(obj *Config) (*Config, error) {
	return n.iface.Create(obj)
}

func (n *configClient2) Get(namespace, name string, opts metav1.GetOptions) (*Config, error) {
	return n.iface.GetNamespaced(namespace, name, opts)
}

func (n *configClient2) Update(obj *Config) (*Config, error) {
	return n.iface.Update(obj)
}

func (n *configClient2) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	return n.iface.DeleteNamespaced(namespace, name, options)
}

func (n *configClient2) List(namespace string, opts metav1.ListOptions) (*ConfigList, error) {
	return n.iface.List(opts)
}

func (n *configClient2) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return n.iface.Watch(opts)
}

func (n *configClientCache) Get(namespace, name string) (*Config, error) {
	return n.client.controller.Lister().Get(namespace, name)
}

func (n *configClientCache) List(namespace string, selector labels.Selector) ([]*Config, error) {
	return n.client.controller.Lister().List(namespace, selector)
}

func (n *configClient2) Cache() ConfigClientCache {
	n.loadController()
	return &configClientCache{
		client: n,
	}
}

func (n *configClient2) OnCreate(ctx context.Context, name string, sync ConfigChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-create", &configLifecycleDelegate{create: sync})
}

func (n *configClient2) OnChange(ctx context.Context, name string, sync ConfigChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-change", &configLifecycleDelegate{update: sync})
}

func (n *configClient2) OnRemove(ctx context.Context, name string, sync ConfigChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name, &configLifecycleDelegate{remove: sync})
}

func (n *configClientCache) Index(name string, indexer ConfigIndexer) {
	err := n.client.controller.Informer().GetIndexer().AddIndexers(map[string]cache.IndexFunc{
		name: func(obj interface{}) ([]string, error) {
			if v, ok := obj.(*Config); ok {
				return indexer(v)
			}
			return nil, nil
		},
	})

	if err != nil {
		panic(err)
	}
}

func (n *configClientCache) GetIndexed(name, key string) ([]*Config, error) {
	var result []*Config
	objs, err := n.client.controller.Informer().GetIndexer().ByIndex(name, key)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		if v, ok := obj.(*Config); ok {
			result = append(result, v)
		}
	}

	return result, nil
}

func (n *configClient2) loadController() {
	if n.controller == nil {
		n.controller = n.iface.Controller()
	}
}

type configLifecycleDelegate struct {
	create ConfigChangeHandlerFunc
	update ConfigChangeHandlerFunc
	remove ConfigChangeHandlerFunc
}

func (n *configLifecycleDelegate) HasCreate() bool {
	return n.create != nil
}

func (n *configLifecycleDelegate) Create(obj *Config) (runtime.Object, error) {
	if n.create == nil {
		return obj, nil
	}
	return n.create(obj)
}

func (n *configLifecycleDelegate) HasFinalize() bool {
	return n.remove != nil
}

func (n *configLifecycleDelegate) Remove(obj *Config) (runtime.Object, error) {
	if n.remove == nil {
		return obj, nil
	}
	return n.remove(obj)
}

func (n *configLifecycleDelegate) Updated(obj *Config) (runtime.Object, error) {
	if n.update == nil {
		return obj, nil
	}
	return n.update(obj)
}
