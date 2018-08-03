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
	VolumeGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "Volume",
	}
	VolumeResource = metav1.APIResource{
		Name:         "volumes",
		SingularName: "volume",
		Namespaced:   true,

		Kind: VolumeGroupVersionKind.Kind,
	}
)

type VolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Volume
}

type VolumeHandlerFunc func(key string, obj *Volume) error

type VolumeLister interface {
	List(namespace string, selector labels.Selector) (ret []*Volume, err error)
	Get(namespace, name string) (*Volume, error)
}

type VolumeController interface {
	Informer() cache.SharedIndexInformer
	Lister() VolumeLister
	AddHandler(name string, handler VolumeHandlerFunc)
	AddClusterScopedHandler(name, clusterName string, handler VolumeHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type VolumeInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*Volume) (*Volume, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*Volume, error)
	Get(name string, opts metav1.GetOptions) (*Volume, error)
	Update(*Volume) (*Volume, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*VolumeList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() VolumeController
	AddHandler(name string, sync VolumeHandlerFunc)
	AddLifecycle(name string, lifecycle VolumeLifecycle)
	AddClusterScopedHandler(name, clusterName string, sync VolumeHandlerFunc)
	AddClusterScopedLifecycle(name, clusterName string, lifecycle VolumeLifecycle)
}

type volumeLister struct {
	controller *volumeController
}

func (l *volumeLister) List(namespace string, selector labels.Selector) (ret []*Volume, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*Volume))
	})
	return
}

func (l *volumeLister) Get(namespace, name string) (*Volume, error) {
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
			Group:    VolumeGroupVersionKind.Group,
			Resource: "volume",
		}, key)
	}
	return obj.(*Volume), nil
}

type volumeController struct {
	controller.GenericController
}

func (c *volumeController) Lister() VolumeLister {
	return &volumeLister{
		controller: c,
	}
}

func (c *volumeController) AddHandler(name string, handler VolumeHandlerFunc) {
	c.GenericController.AddHandler(name, func(key string) error {
		obj, exists, err := c.Informer().GetStore().GetByKey(key)
		if err != nil {
			return err
		}
		if !exists {
			return handler(key, nil)
		}
		return handler(key, obj.(*Volume))
	})
}

func (c *volumeController) AddClusterScopedHandler(name, cluster string, handler VolumeHandlerFunc) {
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

		return handler(key, obj.(*Volume))
	})
}

type volumeFactory struct {
}

func (c volumeFactory) Object() runtime.Object {
	return &Volume{}
}

func (c volumeFactory) List() runtime.Object {
	return &VolumeList{}
}

func (s *volumeClient) Controller() VolumeController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.volumeControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(VolumeGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &volumeController{
		GenericController: genericController,
	}

	s.client.volumeControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type volumeClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   VolumeController
}

func (s *volumeClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *volumeClient) Create(o *Volume) (*Volume, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*Volume), err
}

func (s *volumeClient) Get(name string, opts metav1.GetOptions) (*Volume, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*Volume), err
}

func (s *volumeClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*Volume, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*Volume), err
}

func (s *volumeClient) Update(o *Volume) (*Volume, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*Volume), err
}

func (s *volumeClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *volumeClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *volumeClient) List(opts metav1.ListOptions) (*VolumeList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*VolumeList), err
}

func (s *volumeClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *volumeClient) Patch(o *Volume, data []byte, subresources ...string) (*Volume, error) {
	obj, err := s.objectClient.Patch(o.Name, o, data, subresources...)
	return obj.(*Volume), err
}

func (s *volumeClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *volumeClient) AddHandler(name string, sync VolumeHandlerFunc) {
	s.Controller().AddHandler(name, sync)
}

func (s *volumeClient) AddLifecycle(name string, lifecycle VolumeLifecycle) {
	sync := NewVolumeLifecycleAdapter(name, false, s, lifecycle)
	s.AddHandler(name, sync)
}

func (s *volumeClient) AddClusterScopedHandler(name, clusterName string, sync VolumeHandlerFunc) {
	s.Controller().AddClusterScopedHandler(name, clusterName, sync)
}

func (s *volumeClient) AddClusterScopedLifecycle(name, clusterName string, lifecycle VolumeLifecycle) {
	sync := NewVolumeLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.AddClusterScopedHandler(name, clusterName, sync)
}
