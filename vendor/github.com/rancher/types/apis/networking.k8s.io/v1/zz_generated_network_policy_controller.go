package v1

import (
	"context"

	"github.com/rancher/norman/controller"
	"github.com/rancher/norman/objectclient"
	"k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

var (
	NetworkPolicyGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "NetworkPolicy",
	}
	NetworkPolicyResource = metav1.APIResource{
		Name:         "networkpolicies",
		SingularName: "networkpolicy",
		Namespaced:   true,

		Kind: NetworkPolicyGroupVersionKind.Kind,
	}
)

type NetworkPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []v1.NetworkPolicy
}

type NetworkPolicyHandlerFunc func(key string, obj *v1.NetworkPolicy) error

type NetworkPolicyLister interface {
	List(namespace string, selector labels.Selector) (ret []*v1.NetworkPolicy, err error)
	Get(namespace, name string) (*v1.NetworkPolicy, error)
}

type NetworkPolicyController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() NetworkPolicyLister
	AddHandler(name string, handler NetworkPolicyHandlerFunc)
	AddClusterScopedHandler(name, clusterName string, handler NetworkPolicyHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type NetworkPolicyInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*v1.NetworkPolicy) (*v1.NetworkPolicy, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*v1.NetworkPolicy, error)
	Get(name string, opts metav1.GetOptions) (*v1.NetworkPolicy, error)
	Update(*v1.NetworkPolicy) (*v1.NetworkPolicy, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*NetworkPolicyList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() NetworkPolicyController
	AddHandler(name string, sync NetworkPolicyHandlerFunc)
	AddLifecycle(name string, lifecycle NetworkPolicyLifecycle)
	AddClusterScopedHandler(name, clusterName string, sync NetworkPolicyHandlerFunc)
	AddClusterScopedLifecycle(name, clusterName string, lifecycle NetworkPolicyLifecycle)
}

type networkPolicyLister struct {
	controller *networkPolicyController
}

func (l *networkPolicyLister) List(namespace string, selector labels.Selector) (ret []*v1.NetworkPolicy, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*v1.NetworkPolicy))
	})
	return
}

func (l *networkPolicyLister) Get(namespace, name string) (*v1.NetworkPolicy, error) {
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
			Group:    NetworkPolicyGroupVersionKind.Group,
			Resource: "networkPolicy",
		}, key)
	}
	return obj.(*v1.NetworkPolicy), nil
}

type networkPolicyController struct {
	controller.GenericController
}

func (c *networkPolicyController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *networkPolicyController) Lister() NetworkPolicyLister {
	return &networkPolicyLister{
		controller: c,
	}
}

func (c *networkPolicyController) AddHandler(name string, handler NetworkPolicyHandlerFunc) {
	c.GenericController.AddHandler(name, func(key string) error {
		obj, exists, err := c.Informer().GetStore().GetByKey(key)
		if err != nil {
			return err
		}
		if !exists {
			return handler(key, nil)
		}
		return handler(key, obj.(*v1.NetworkPolicy))
	})
}

func (c *networkPolicyController) AddClusterScopedHandler(name, cluster string, handler NetworkPolicyHandlerFunc) {
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

		return handler(key, obj.(*v1.NetworkPolicy))
	})
}

type networkPolicyFactory struct {
}

func (c networkPolicyFactory) Object() runtime.Object {
	return &v1.NetworkPolicy{}
}

func (c networkPolicyFactory) List() runtime.Object {
	return &NetworkPolicyList{}
}

func (s *networkPolicyClient) Controller() NetworkPolicyController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.networkPolicyControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(NetworkPolicyGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &networkPolicyController{
		GenericController: genericController,
	}

	s.client.networkPolicyControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type networkPolicyClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   NetworkPolicyController
}

func (s *networkPolicyClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *networkPolicyClient) Create(o *v1.NetworkPolicy) (*v1.NetworkPolicy, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*v1.NetworkPolicy), err
}

func (s *networkPolicyClient) Get(name string, opts metav1.GetOptions) (*v1.NetworkPolicy, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*v1.NetworkPolicy), err
}

func (s *networkPolicyClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*v1.NetworkPolicy, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*v1.NetworkPolicy), err
}

func (s *networkPolicyClient) Update(o *v1.NetworkPolicy) (*v1.NetworkPolicy, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*v1.NetworkPolicy), err
}

func (s *networkPolicyClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *networkPolicyClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *networkPolicyClient) List(opts metav1.ListOptions) (*NetworkPolicyList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*NetworkPolicyList), err
}

func (s *networkPolicyClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *networkPolicyClient) Patch(o *v1.NetworkPolicy, data []byte, subresources ...string) (*v1.NetworkPolicy, error) {
	obj, err := s.objectClient.Patch(o.Name, o, data, subresources...)
	return obj.(*v1.NetworkPolicy), err
}

func (s *networkPolicyClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *networkPolicyClient) AddHandler(name string, sync NetworkPolicyHandlerFunc) {
	s.Controller().AddHandler(name, sync)
}

func (s *networkPolicyClient) AddLifecycle(name string, lifecycle NetworkPolicyLifecycle) {
	sync := NewNetworkPolicyLifecycleAdapter(name, false, s, lifecycle)
	s.AddHandler(name, sync)
}

func (s *networkPolicyClient) AddClusterScopedHandler(name, clusterName string, sync NetworkPolicyHandlerFunc) {
	s.Controller().AddClusterScopedHandler(name, clusterName, sync)
}

func (s *networkPolicyClient) AddClusterScopedLifecycle(name, clusterName string, lifecycle NetworkPolicyLifecycle) {
	sync := NewNetworkPolicyLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.AddClusterScopedHandler(name, clusterName, sync)
}
