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
	PipelineSettingGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "PipelineSetting",
	}
	PipelineSettingResource = metav1.APIResource{
		Name:         "pipelinesettings",
		SingularName: "pipelinesetting",
		Namespaced:   true,

		Kind: PipelineSettingGroupVersionKind.Kind,
	}
)

type PipelineSettingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PipelineSetting
}

type PipelineSettingHandlerFunc func(key string, obj *PipelineSetting) error

type PipelineSettingLister interface {
	List(namespace string, selector labels.Selector) (ret []*PipelineSetting, err error)
	Get(namespace, name string) (*PipelineSetting, error)
}

type PipelineSettingController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() PipelineSettingLister
	AddHandler(name string, handler PipelineSettingHandlerFunc)
	AddClusterScopedHandler(name, clusterName string, handler PipelineSettingHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type PipelineSettingInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*PipelineSetting) (*PipelineSetting, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*PipelineSetting, error)
	Get(name string, opts metav1.GetOptions) (*PipelineSetting, error)
	Update(*PipelineSetting) (*PipelineSetting, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*PipelineSettingList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() PipelineSettingController
	AddHandler(name string, sync PipelineSettingHandlerFunc)
	AddLifecycle(name string, lifecycle PipelineSettingLifecycle)
	AddClusterScopedHandler(name, clusterName string, sync PipelineSettingHandlerFunc)
	AddClusterScopedLifecycle(name, clusterName string, lifecycle PipelineSettingLifecycle)
}

type pipelineSettingLister struct {
	controller *pipelineSettingController
}

func (l *pipelineSettingLister) List(namespace string, selector labels.Selector) (ret []*PipelineSetting, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*PipelineSetting))
	})
	return
}

func (l *pipelineSettingLister) Get(namespace, name string) (*PipelineSetting, error) {
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
			Group:    PipelineSettingGroupVersionKind.Group,
			Resource: "pipelineSetting",
		}, key)
	}
	return obj.(*PipelineSetting), nil
}

type pipelineSettingController struct {
	controller.GenericController
}

func (c *pipelineSettingController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *pipelineSettingController) Lister() PipelineSettingLister {
	return &pipelineSettingLister{
		controller: c,
	}
}

func (c *pipelineSettingController) AddHandler(name string, handler PipelineSettingHandlerFunc) {
	c.GenericController.AddHandler(name, func(key string) error {
		obj, exists, err := c.Informer().GetStore().GetByKey(key)
		if err != nil {
			return err
		}
		if !exists {
			return handler(key, nil)
		}
		return handler(key, obj.(*PipelineSetting))
	})
}

func (c *pipelineSettingController) AddClusterScopedHandler(name, cluster string, handler PipelineSettingHandlerFunc) {
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

		return handler(key, obj.(*PipelineSetting))
	})
}

type pipelineSettingFactory struct {
}

func (c pipelineSettingFactory) Object() runtime.Object {
	return &PipelineSetting{}
}

func (c pipelineSettingFactory) List() runtime.Object {
	return &PipelineSettingList{}
}

func (s *pipelineSettingClient) Controller() PipelineSettingController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.pipelineSettingControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(PipelineSettingGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &pipelineSettingController{
		GenericController: genericController,
	}

	s.client.pipelineSettingControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type pipelineSettingClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   PipelineSettingController
}

func (s *pipelineSettingClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *pipelineSettingClient) Create(o *PipelineSetting) (*PipelineSetting, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*PipelineSetting), err
}

func (s *pipelineSettingClient) Get(name string, opts metav1.GetOptions) (*PipelineSetting, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*PipelineSetting), err
}

func (s *pipelineSettingClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*PipelineSetting, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*PipelineSetting), err
}

func (s *pipelineSettingClient) Update(o *PipelineSetting) (*PipelineSetting, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*PipelineSetting), err
}

func (s *pipelineSettingClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *pipelineSettingClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *pipelineSettingClient) List(opts metav1.ListOptions) (*PipelineSettingList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*PipelineSettingList), err
}

func (s *pipelineSettingClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *pipelineSettingClient) Patch(o *PipelineSetting, data []byte, subresources ...string) (*PipelineSetting, error) {
	obj, err := s.objectClient.Patch(o.Name, o, data, subresources...)
	return obj.(*PipelineSetting), err
}

func (s *pipelineSettingClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *pipelineSettingClient) AddHandler(name string, sync PipelineSettingHandlerFunc) {
	s.Controller().AddHandler(name, sync)
}

func (s *pipelineSettingClient) AddLifecycle(name string, lifecycle PipelineSettingLifecycle) {
	sync := NewPipelineSettingLifecycleAdapter(name, false, s, lifecycle)
	s.AddHandler(name, sync)
}

func (s *pipelineSettingClient) AddClusterScopedHandler(name, clusterName string, sync PipelineSettingHandlerFunc) {
	s.Controller().AddClusterScopedHandler(name, clusterName, sync)
}

func (s *pipelineSettingClient) AddClusterScopedLifecycle(name, clusterName string, lifecycle PipelineSettingLifecycle) {
	sync := NewPipelineSettingLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.AddClusterScopedHandler(name, clusterName, sync)
}
