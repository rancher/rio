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
	RouteSetGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "RouteSet",
	}
	RouteSetResource = metav1.APIResource{
		Name:         "routesets",
		SingularName: "routeset",
		Namespaced:   true,

		Kind: RouteSetGroupVersionKind.Kind,
	}
)

type RouteSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RouteSet
}

type RouteSetHandlerFunc func(key string, obj *RouteSet) error

type RouteSetLister interface {
	List(namespace string, selector labels.Selector) (ret []*RouteSet, err error)
	Get(namespace, name string) (*RouteSet, error)
}

type RouteSetController interface {
	Informer() cache.SharedIndexInformer
	Lister() RouteSetLister
	AddHandler(name string, handler RouteSetHandlerFunc)
	AddClusterScopedHandler(name, clusterName string, handler RouteSetHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type RouteSetInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*RouteSet) (*RouteSet, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*RouteSet, error)
	Get(name string, opts metav1.GetOptions) (*RouteSet, error)
	Update(*RouteSet) (*RouteSet, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*RouteSetList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() RouteSetController
	AddHandler(name string, sync RouteSetHandlerFunc)
	AddLifecycle(name string, lifecycle RouteSetLifecycle)
	AddClusterScopedHandler(name, clusterName string, sync RouteSetHandlerFunc)
	AddClusterScopedLifecycle(name, clusterName string, lifecycle RouteSetLifecycle)
}

type routeSetLister struct {
	controller *routeSetController
}

func (l *routeSetLister) List(namespace string, selector labels.Selector) (ret []*RouteSet, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*RouteSet))
	})
	return
}

func (l *routeSetLister) Get(namespace, name string) (*RouteSet, error) {
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
			Group:    RouteSetGroupVersionKind.Group,
			Resource: "routeSet",
		}, key)
	}
	return obj.(*RouteSet), nil
}

type routeSetController struct {
	controller.GenericController
}

func (c *routeSetController) Lister() RouteSetLister {
	return &routeSetLister{
		controller: c,
	}
}

func (c *routeSetController) AddHandler(name string, handler RouteSetHandlerFunc) {
	c.GenericController.AddHandler(name, func(key string) error {
		obj, exists, err := c.Informer().GetStore().GetByKey(key)
		if err != nil {
			return err
		}
		if !exists {
			return handler(key, nil)
		}
		return handler(key, obj.(*RouteSet))
	})
}

func (c *routeSetController) AddClusterScopedHandler(name, cluster string, handler RouteSetHandlerFunc) {
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

		return handler(key, obj.(*RouteSet))
	})
}

type routeSetFactory struct {
}

func (c routeSetFactory) Object() runtime.Object {
	return &RouteSet{}
}

func (c routeSetFactory) List() runtime.Object {
	return &RouteSetList{}
}

func (s *routeSetClient) Controller() RouteSetController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.routeSetControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(RouteSetGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &routeSetController{
		GenericController: genericController,
	}

	s.client.routeSetControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type routeSetClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   RouteSetController
}

func (s *routeSetClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *routeSetClient) Create(o *RouteSet) (*RouteSet, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*RouteSet), err
}

func (s *routeSetClient) Get(name string, opts metav1.GetOptions) (*RouteSet, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*RouteSet), err
}

func (s *routeSetClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*RouteSet, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*RouteSet), err
}

func (s *routeSetClient) Update(o *RouteSet) (*RouteSet, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*RouteSet), err
}

func (s *routeSetClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *routeSetClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *routeSetClient) List(opts metav1.ListOptions) (*RouteSetList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*RouteSetList), err
}

func (s *routeSetClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *routeSetClient) Patch(o *RouteSet, data []byte, subresources ...string) (*RouteSet, error) {
	obj, err := s.objectClient.Patch(o.Name, o, data, subresources...)
	return obj.(*RouteSet), err
}

func (s *routeSetClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *routeSetClient) AddHandler(name string, sync RouteSetHandlerFunc) {
	s.Controller().AddHandler(name, sync)
}

func (s *routeSetClient) AddLifecycle(name string, lifecycle RouteSetLifecycle) {
	sync := NewRouteSetLifecycleAdapter(name, false, s, lifecycle)
	s.AddHandler(name, sync)
}

func (s *routeSetClient) AddClusterScopedHandler(name, clusterName string, sync RouteSetHandlerFunc) {
	s.Controller().AddClusterScopedHandler(name, clusterName, sync)
}

func (s *routeSetClient) AddClusterScopedLifecycle(name, clusterName string, lifecycle RouteSetLifecycle) {
	sync := NewRouteSetLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.AddClusterScopedHandler(name, clusterName, sync)
}
