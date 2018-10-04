package v1

import (
	"context"

	"github.com/rancher/norman/controller"
	"github.com/rancher/norman/objectclient"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

var (
	ConfigMapGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "ConfigMap",
	}
	ConfigMapResource = metav1.APIResource{
		Name:         "configmaps",
		SingularName: "configmap",
		Namespaced:   true,

		Kind: ConfigMapGroupVersionKind.Kind,
	}
)

type ConfigMapList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []v1.ConfigMap
}

type ConfigMapHandlerFunc func(key string, obj *v1.ConfigMap) error

type ConfigMapLister interface {
	List(namespace string, selector labels.Selector) (ret []*v1.ConfigMap, err error)
	Get(namespace, name string) (*v1.ConfigMap, error)
}

type ConfigMapController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() ConfigMapLister
	AddHandler(name string, handler ConfigMapHandlerFunc)
	AddClusterScopedHandler(name, clusterName string, handler ConfigMapHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type ConfigMapInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*v1.ConfigMap) (*v1.ConfigMap, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*v1.ConfigMap, error)
	Get(name string, opts metav1.GetOptions) (*v1.ConfigMap, error)
	Update(*v1.ConfigMap) (*v1.ConfigMap, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*ConfigMapList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() ConfigMapController
	AddHandler(name string, sync ConfigMapHandlerFunc)
	AddLifecycle(name string, lifecycle ConfigMapLifecycle)
	AddClusterScopedHandler(name, clusterName string, sync ConfigMapHandlerFunc)
	AddClusterScopedLifecycle(name, clusterName string, lifecycle ConfigMapLifecycle)
}

type configMapLister struct {
	controller *configMapController
}

func (l *configMapLister) List(namespace string, selector labels.Selector) (ret []*v1.ConfigMap, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*v1.ConfigMap))
	})
	return
}

func (l *configMapLister) Get(namespace, name string) (*v1.ConfigMap, error) {
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
			Group:    ConfigMapGroupVersionKind.Group,
			Resource: "configMap",
		}, key)
	}
	return obj.(*v1.ConfigMap), nil
}

type configMapController struct {
	controller.GenericController
}

func (c *configMapController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *configMapController) Lister() ConfigMapLister {
	return &configMapLister{
		controller: c,
	}
}

func (c *configMapController) AddHandler(name string, handler ConfigMapHandlerFunc) {
	c.GenericController.AddHandler(name, func(key string) error {
		obj, exists, err := c.Informer().GetStore().GetByKey(key)
		if err != nil {
			return err
		}
		if !exists {
			return handler(key, nil)
		}
		return handler(key, obj.(*v1.ConfigMap))
	})
}

func (c *configMapController) AddClusterScopedHandler(name, cluster string, handler ConfigMapHandlerFunc) {
	c.GenericController.AddHandler(name, func(key string) error {
		obj, exists, err := c.Informer().GetStore().GetByKey(key)
		if err != nil {
			return err
		}
		if !exists {
			return handler(key, nil)
		}

		if !controller.ObjectInCluster(cluster, obj) {
			return nil
		}

		return handler(key, obj.(*v1.ConfigMap))
	})
}

type configMapFactory struct {
}

func (c configMapFactory) Object() runtime.Object {
	return &v1.ConfigMap{}
}

func (c configMapFactory) List() runtime.Object {
	return &ConfigMapList{}
}

func (s *configMapClient) Controller() ConfigMapController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.configMapControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(ConfigMapGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &configMapController{
		GenericController: genericController,
	}

	s.client.configMapControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type configMapClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   ConfigMapController
}

func (s *configMapClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *configMapClient) Create(o *v1.ConfigMap) (*v1.ConfigMap, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*v1.ConfigMap), err
}

func (s *configMapClient) Get(name string, opts metav1.GetOptions) (*v1.ConfigMap, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*v1.ConfigMap), err
}

func (s *configMapClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*v1.ConfigMap, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*v1.ConfigMap), err
}

func (s *configMapClient) Update(o *v1.ConfigMap) (*v1.ConfigMap, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*v1.ConfigMap), err
}

func (s *configMapClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *configMapClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *configMapClient) List(opts metav1.ListOptions) (*ConfigMapList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*ConfigMapList), err
}

func (s *configMapClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *configMapClient) Patch(o *v1.ConfigMap, data []byte, subresources ...string) (*v1.ConfigMap, error) {
	obj, err := s.objectClient.Patch(o.Name, o, data, subresources...)
	return obj.(*v1.ConfigMap), err
}

func (s *configMapClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *configMapClient) AddHandler(name string, sync ConfigMapHandlerFunc) {
	s.Controller().AddHandler(name, sync)
}

func (s *configMapClient) AddLifecycle(name string, lifecycle ConfigMapLifecycle) {
	sync := NewConfigMapLifecycleAdapter(name, false, s, lifecycle)
	s.AddHandler(name, sync)
}

func (s *configMapClient) AddClusterScopedHandler(name, clusterName string, sync ConfigMapHandlerFunc) {
	s.Controller().AddClusterScopedHandler(name, clusterName, sync)
}

func (s *configMapClient) AddClusterScopedLifecycle(name, clusterName string, lifecycle ConfigMapLifecycle) {
	sync := NewConfigMapLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.AddClusterScopedHandler(name, clusterName, sync)
}
