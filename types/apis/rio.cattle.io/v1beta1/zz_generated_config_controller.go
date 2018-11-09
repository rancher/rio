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

type ConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Config
}

type ConfigHandlerFunc func(key string, obj *Config) (runtime.Object, error)

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
func (s *configClient) Patch(o *Config, data []byte, subresources ...string) (*Config, error) {
	obj, err := s.objectClient.Patch(o.Name, o, data, subresources...)
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
