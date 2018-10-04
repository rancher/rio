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
	ComponentStatusGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "ComponentStatus",
	}
	ComponentStatusResource = metav1.APIResource{
		Name:         "componentstatuses",
		SingularName: "componentstatus",
		Namespaced:   false,
		Kind:         ComponentStatusGroupVersionKind.Kind,
	}
)

type ComponentStatusList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []v1.ComponentStatus
}

type ComponentStatusHandlerFunc func(key string, obj *v1.ComponentStatus) error

type ComponentStatusLister interface {
	List(namespace string, selector labels.Selector) (ret []*v1.ComponentStatus, err error)
	Get(namespace, name string) (*v1.ComponentStatus, error)
}

type ComponentStatusController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() ComponentStatusLister
	AddHandler(name string, handler ComponentStatusHandlerFunc)
	AddClusterScopedHandler(name, clusterName string, handler ComponentStatusHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type ComponentStatusInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*v1.ComponentStatus) (*v1.ComponentStatus, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*v1.ComponentStatus, error)
	Get(name string, opts metav1.GetOptions) (*v1.ComponentStatus, error)
	Update(*v1.ComponentStatus) (*v1.ComponentStatus, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*ComponentStatusList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() ComponentStatusController
	AddHandler(name string, sync ComponentStatusHandlerFunc)
	AddLifecycle(name string, lifecycle ComponentStatusLifecycle)
	AddClusterScopedHandler(name, clusterName string, sync ComponentStatusHandlerFunc)
	AddClusterScopedLifecycle(name, clusterName string, lifecycle ComponentStatusLifecycle)
}

type componentStatusLister struct {
	controller *componentStatusController
}

func (l *componentStatusLister) List(namespace string, selector labels.Selector) (ret []*v1.ComponentStatus, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*v1.ComponentStatus))
	})
	return
}

func (l *componentStatusLister) Get(namespace, name string) (*v1.ComponentStatus, error) {
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
			Group:    ComponentStatusGroupVersionKind.Group,
			Resource: "componentStatus",
		}, key)
	}
	return obj.(*v1.ComponentStatus), nil
}

type componentStatusController struct {
	controller.GenericController
}

func (c *componentStatusController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *componentStatusController) Lister() ComponentStatusLister {
	return &componentStatusLister{
		controller: c,
	}
}

func (c *componentStatusController) AddHandler(name string, handler ComponentStatusHandlerFunc) {
	c.GenericController.AddHandler(name, func(key string) error {
		obj, exists, err := c.Informer().GetStore().GetByKey(key)
		if err != nil {
			return err
		}
		if !exists {
			return handler(key, nil)
		}
		return handler(key, obj.(*v1.ComponentStatus))
	})
}

func (c *componentStatusController) AddClusterScopedHandler(name, cluster string, handler ComponentStatusHandlerFunc) {
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

		return handler(key, obj.(*v1.ComponentStatus))
	})
}

type componentStatusFactory struct {
}

func (c componentStatusFactory) Object() runtime.Object {
	return &v1.ComponentStatus{}
}

func (c componentStatusFactory) List() runtime.Object {
	return &ComponentStatusList{}
}

func (s *componentStatusClient) Controller() ComponentStatusController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.componentStatusControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(ComponentStatusGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &componentStatusController{
		GenericController: genericController,
	}

	s.client.componentStatusControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type componentStatusClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   ComponentStatusController
}

func (s *componentStatusClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *componentStatusClient) Create(o *v1.ComponentStatus) (*v1.ComponentStatus, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*v1.ComponentStatus), err
}

func (s *componentStatusClient) Get(name string, opts metav1.GetOptions) (*v1.ComponentStatus, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*v1.ComponentStatus), err
}

func (s *componentStatusClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*v1.ComponentStatus, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*v1.ComponentStatus), err
}

func (s *componentStatusClient) Update(o *v1.ComponentStatus) (*v1.ComponentStatus, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*v1.ComponentStatus), err
}

func (s *componentStatusClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *componentStatusClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *componentStatusClient) List(opts metav1.ListOptions) (*ComponentStatusList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*ComponentStatusList), err
}

func (s *componentStatusClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *componentStatusClient) Patch(o *v1.ComponentStatus, data []byte, subresources ...string) (*v1.ComponentStatus, error) {
	obj, err := s.objectClient.Patch(o.Name, o, data, subresources...)
	return obj.(*v1.ComponentStatus), err
}

func (s *componentStatusClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *componentStatusClient) AddHandler(name string, sync ComponentStatusHandlerFunc) {
	s.Controller().AddHandler(name, sync)
}

func (s *componentStatusClient) AddLifecycle(name string, lifecycle ComponentStatusLifecycle) {
	sync := NewComponentStatusLifecycleAdapter(name, false, s, lifecycle)
	s.AddHandler(name, sync)
}

func (s *componentStatusClient) AddClusterScopedHandler(name, clusterName string, sync ComponentStatusHandlerFunc) {
	s.Controller().AddClusterScopedHandler(name, clusterName, sync)
}

func (s *componentStatusClient) AddClusterScopedLifecycle(name, clusterName string, lifecycle ComponentStatusLifecycle) {
	sync := NewComponentStatusLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.AddClusterScopedHandler(name, clusterName, sync)
}
