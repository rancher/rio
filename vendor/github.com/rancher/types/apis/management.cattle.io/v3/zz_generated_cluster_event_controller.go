package v3

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
	ClusterEventGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "ClusterEvent",
	}
	ClusterEventResource = metav1.APIResource{
		Name:         "clusterevents",
		SingularName: "clusterevent",
		Namespaced:   true,

		Kind: ClusterEventGroupVersionKind.Kind,
	}
)

type ClusterEventList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterEvent
}

type ClusterEventHandlerFunc func(key string, obj *ClusterEvent) error

type ClusterEventLister interface {
	List(namespace string, selector labels.Selector) (ret []*ClusterEvent, err error)
	Get(namespace, name string) (*ClusterEvent, error)
}

type ClusterEventController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() ClusterEventLister
	AddHandler(name string, handler ClusterEventHandlerFunc)
	AddClusterScopedHandler(name, clusterName string, handler ClusterEventHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type ClusterEventInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*ClusterEvent) (*ClusterEvent, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*ClusterEvent, error)
	Get(name string, opts metav1.GetOptions) (*ClusterEvent, error)
	Update(*ClusterEvent) (*ClusterEvent, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*ClusterEventList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() ClusterEventController
	AddHandler(name string, sync ClusterEventHandlerFunc)
	AddLifecycle(name string, lifecycle ClusterEventLifecycle)
	AddClusterScopedHandler(name, clusterName string, sync ClusterEventHandlerFunc)
	AddClusterScopedLifecycle(name, clusterName string, lifecycle ClusterEventLifecycle)
}

type clusterEventLister struct {
	controller *clusterEventController
}

func (l *clusterEventLister) List(namespace string, selector labels.Selector) (ret []*ClusterEvent, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*ClusterEvent))
	})
	return
}

func (l *clusterEventLister) Get(namespace, name string) (*ClusterEvent, error) {
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
			Group:    ClusterEventGroupVersionKind.Group,
			Resource: "clusterEvent",
		}, key)
	}
	return obj.(*ClusterEvent), nil
}

type clusterEventController struct {
	controller.GenericController
}

func (c *clusterEventController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *clusterEventController) Lister() ClusterEventLister {
	return &clusterEventLister{
		controller: c,
	}
}

func (c *clusterEventController) AddHandler(name string, handler ClusterEventHandlerFunc) {
	c.GenericController.AddHandler(name, func(key string) error {
		obj, exists, err := c.Informer().GetStore().GetByKey(key)
		if err != nil {
			return err
		}
		if !exists {
			return handler(key, nil)
		}
		return handler(key, obj.(*ClusterEvent))
	})
}

func (c *clusterEventController) AddClusterScopedHandler(name, cluster string, handler ClusterEventHandlerFunc) {
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

		return handler(key, obj.(*ClusterEvent))
	})
}

type clusterEventFactory struct {
}

func (c clusterEventFactory) Object() runtime.Object {
	return &ClusterEvent{}
}

func (c clusterEventFactory) List() runtime.Object {
	return &ClusterEventList{}
}

func (s *clusterEventClient) Controller() ClusterEventController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.clusterEventControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(ClusterEventGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &clusterEventController{
		GenericController: genericController,
	}

	s.client.clusterEventControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type clusterEventClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   ClusterEventController
}

func (s *clusterEventClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *clusterEventClient) Create(o *ClusterEvent) (*ClusterEvent, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*ClusterEvent), err
}

func (s *clusterEventClient) Get(name string, opts metav1.GetOptions) (*ClusterEvent, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*ClusterEvent), err
}

func (s *clusterEventClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*ClusterEvent, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*ClusterEvent), err
}

func (s *clusterEventClient) Update(o *ClusterEvent) (*ClusterEvent, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*ClusterEvent), err
}

func (s *clusterEventClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *clusterEventClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *clusterEventClient) List(opts metav1.ListOptions) (*ClusterEventList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*ClusterEventList), err
}

func (s *clusterEventClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *clusterEventClient) Patch(o *ClusterEvent, data []byte, subresources ...string) (*ClusterEvent, error) {
	obj, err := s.objectClient.Patch(o.Name, o, data, subresources...)
	return obj.(*ClusterEvent), err
}

func (s *clusterEventClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *clusterEventClient) AddHandler(name string, sync ClusterEventHandlerFunc) {
	s.Controller().AddHandler(name, sync)
}

func (s *clusterEventClient) AddLifecycle(name string, lifecycle ClusterEventLifecycle) {
	sync := NewClusterEventLifecycleAdapter(name, false, s, lifecycle)
	s.AddHandler(name, sync)
}

func (s *clusterEventClient) AddClusterScopedHandler(name, clusterName string, sync ClusterEventHandlerFunc) {
	s.Controller().AddClusterScopedHandler(name, clusterName, sync)
}

func (s *clusterEventClient) AddClusterScopedLifecycle(name, clusterName string, lifecycle ClusterEventLifecycle) {
	sync := NewClusterEventLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.AddClusterScopedHandler(name, clusterName, sync)
}
