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
	PublicDomainGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "PublicDomain",
	}
	PublicDomainResource = metav1.APIResource{
		Name:         "publicdomains",
		SingularName: "publicdomain",
		Namespaced:   false,
		Kind:         PublicDomainGroupVersionKind.Kind,
	}
)

type PublicDomainList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PublicDomain
}

type PublicDomainHandlerFunc func(key string, obj *PublicDomain) error

type PublicDomainLister interface {
	List(namespace string, selector labels.Selector) (ret []*PublicDomain, err error)
	Get(namespace, name string) (*PublicDomain, error)
}

type PublicDomainController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() PublicDomainLister
	AddHandler(name string, handler PublicDomainHandlerFunc)
	AddClusterScopedHandler(name, clusterName string, handler PublicDomainHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type PublicDomainInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*PublicDomain) (*PublicDomain, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*PublicDomain, error)
	Get(name string, opts metav1.GetOptions) (*PublicDomain, error)
	Update(*PublicDomain) (*PublicDomain, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*PublicDomainList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() PublicDomainController
	AddHandler(name string, sync PublicDomainHandlerFunc)
	AddLifecycle(name string, lifecycle PublicDomainLifecycle)
	AddClusterScopedHandler(name, clusterName string, sync PublicDomainHandlerFunc)
	AddClusterScopedLifecycle(name, clusterName string, lifecycle PublicDomainLifecycle)
}

type publicDomainLister struct {
	controller *publicDomainController
}

func (l *publicDomainLister) List(namespace string, selector labels.Selector) (ret []*PublicDomain, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*PublicDomain))
	})
	return
}

func (l *publicDomainLister) Get(namespace, name string) (*PublicDomain, error) {
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
			Group:    PublicDomainGroupVersionKind.Group,
			Resource: "publicDomain",
		}, key)
	}
	return obj.(*PublicDomain), nil
}

type publicDomainController struct {
	controller.GenericController
}

func (c *publicDomainController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *publicDomainController) Lister() PublicDomainLister {
	return &publicDomainLister{
		controller: c,
	}
}

func (c *publicDomainController) AddHandler(name string, handler PublicDomainHandlerFunc) {
	c.GenericController.AddHandler(name, func(key string) error {
		obj, exists, err := c.Informer().GetStore().GetByKey(key)
		if err != nil {
			return err
		}
		if !exists {
			return handler(key, nil)
		}
		return handler(key, obj.(*PublicDomain))
	})
}

func (c *publicDomainController) AddClusterScopedHandler(name, cluster string, handler PublicDomainHandlerFunc) {
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

		return handler(key, obj.(*PublicDomain))
	})
}

type publicDomainFactory struct {
}

func (c publicDomainFactory) Object() runtime.Object {
	return &PublicDomain{}
}

func (c publicDomainFactory) List() runtime.Object {
	return &PublicDomainList{}
}

func (s *publicDomainClient) Controller() PublicDomainController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.publicDomainControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(PublicDomainGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &publicDomainController{
		GenericController: genericController,
	}

	s.client.publicDomainControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type publicDomainClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   PublicDomainController
}

func (s *publicDomainClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *publicDomainClient) Create(o *PublicDomain) (*PublicDomain, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*PublicDomain), err
}

func (s *publicDomainClient) Get(name string, opts metav1.GetOptions) (*PublicDomain, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*PublicDomain), err
}

func (s *publicDomainClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*PublicDomain, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*PublicDomain), err
}

func (s *publicDomainClient) Update(o *PublicDomain) (*PublicDomain, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*PublicDomain), err
}

func (s *publicDomainClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *publicDomainClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *publicDomainClient) List(opts metav1.ListOptions) (*PublicDomainList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*PublicDomainList), err
}

func (s *publicDomainClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *publicDomainClient) Patch(o *PublicDomain, data []byte, subresources ...string) (*PublicDomain, error) {
	obj, err := s.objectClient.Patch(o.Name, o, data, subresources...)
	return obj.(*PublicDomain), err
}

func (s *publicDomainClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *publicDomainClient) AddHandler(name string, sync PublicDomainHandlerFunc) {
	s.Controller().AddHandler(name, sync)
}

func (s *publicDomainClient) AddLifecycle(name string, lifecycle PublicDomainLifecycle) {
	sync := NewPublicDomainLifecycleAdapter(name, false, s, lifecycle)
	s.AddHandler(name, sync)
}

func (s *publicDomainClient) AddClusterScopedHandler(name, clusterName string, sync PublicDomainHandlerFunc) {
	s.Controller().AddClusterScopedHandler(name, clusterName, sync)
}

func (s *publicDomainClient) AddClusterScopedLifecycle(name, clusterName string, lifecycle PublicDomainLifecycle) {
	sync := NewPublicDomainLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.AddClusterScopedHandler(name, clusterName, sync)
}
