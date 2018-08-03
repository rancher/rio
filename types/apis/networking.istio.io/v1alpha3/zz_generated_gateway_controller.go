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
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

var (
	GatewayGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "Gateway",
	}
	GatewayResource = metav1.APIResource{
		Name:         "gateways",
		SingularName: "gateway",
		Namespaced:   true,

		Kind: GatewayGroupVersionKind.Kind,
	}
)

type GatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Gateway
}

type GatewayHandlerFunc func(key string, obj *Gateway) error

type GatewayLister interface {
	List(namespace string, selector labels.Selector) (ret []*Gateway, err error)
	Get(namespace, name string) (*Gateway, error)
}

type GatewayController interface {
	Informer() cache.SharedIndexInformer
	Lister() GatewayLister
	AddHandler(name string, handler GatewayHandlerFunc)
	AddClusterScopedHandler(name, clusterName string, handler GatewayHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type GatewayInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*Gateway) (*Gateway, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*Gateway, error)
	Get(name string, opts metav1.GetOptions) (*Gateway, error)
	Update(*Gateway) (*Gateway, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*GatewayList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() GatewayController
	AddHandler(name string, sync GatewayHandlerFunc)
	AddLifecycle(name string, lifecycle GatewayLifecycle)
	AddClusterScopedHandler(name, clusterName string, sync GatewayHandlerFunc)
	AddClusterScopedLifecycle(name, clusterName string, lifecycle GatewayLifecycle)
}

type gatewayLister struct {
	controller *gatewayController
}

func (l *gatewayLister) List(namespace string, selector labels.Selector) (ret []*Gateway, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*Gateway))
	})
	return
}

func (l *gatewayLister) Get(namespace, name string) (*Gateway, error) {
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
			Group:    GatewayGroupVersionKind.Group,
			Resource: "gateway",
		}, key)
	}
	return obj.(*Gateway), nil
}

type gatewayController struct {
	controller.GenericController
}

func (c *gatewayController) Lister() GatewayLister {
	return &gatewayLister{
		controller: c,
	}
}

func (c *gatewayController) AddHandler(name string, handler GatewayHandlerFunc) {
	c.GenericController.AddHandler(name, func(key string) error {
		obj, exists, err := c.Informer().GetStore().GetByKey(key)
		if err != nil {
			return err
		}
		if !exists {
			return handler(key, nil)
		}
		return handler(key, obj.(*Gateway))
	})
}

func (c *gatewayController) AddClusterScopedHandler(name, cluster string, handler GatewayHandlerFunc) {
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

		return handler(key, obj.(*Gateway))
	})
}

type gatewayFactory struct {
}

func (c gatewayFactory) Object() runtime.Object {
	return &Gateway{}
}

func (c gatewayFactory) List() runtime.Object {
	return &GatewayList{}
}

func (s *gatewayClient) Controller() GatewayController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.gatewayControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(GatewayGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &gatewayController{
		GenericController: genericController,
	}

	s.client.gatewayControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type gatewayClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   GatewayController
}

func (s *gatewayClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *gatewayClient) Create(o *Gateway) (*Gateway, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*Gateway), err
}

func (s *gatewayClient) Get(name string, opts metav1.GetOptions) (*Gateway, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*Gateway), err
}

func (s *gatewayClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*Gateway, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*Gateway), err
}

func (s *gatewayClient) Update(o *Gateway) (*Gateway, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*Gateway), err
}

func (s *gatewayClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *gatewayClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *gatewayClient) List(opts metav1.ListOptions) (*GatewayList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*GatewayList), err
}

func (s *gatewayClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *gatewayClient) Patch(o *Gateway, data []byte, subresources ...string) (*Gateway, error) {
	obj, err := s.objectClient.Patch(o.Name, o, data, subresources...)
	return obj.(*Gateway), err
}

func (s *gatewayClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *gatewayClient) AddHandler(name string, sync GatewayHandlerFunc) {
	s.Controller().AddHandler(name, sync)
}

func (s *gatewayClient) AddLifecycle(name string, lifecycle GatewayLifecycle) {
	sync := NewGatewayLifecycleAdapter(name, false, s, lifecycle)
	s.AddHandler(name, sync)
}

func (s *gatewayClient) AddClusterScopedHandler(name, clusterName string, sync GatewayHandlerFunc) {
	s.Controller().AddClusterScopedHandler(name, clusterName, sync)
}

func (s *gatewayClient) AddClusterScopedLifecycle(name, clusterName string, lifecycle GatewayLifecycle) {
	sync := NewGatewayLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.AddClusterScopedHandler(name, clusterName, sync)
}
