package v1

import (
	"context"

	"github.com/rancher/norman/controller"
	"github.com/rancher/norman/objectclient"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
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

func NewVolume(namespace, name string, obj Volume) *Volume {
	obj.APIVersion, obj.Kind = VolumeGroupVersionKind.ToAPIVersionAndKind()
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

type VolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Volume
}

type VolumeHandlerFunc func(key string, obj *Volume) (runtime.Object, error)

type VolumeChangeHandlerFunc func(obj *Volume) (runtime.Object, error)

type VolumeLister interface {
	List(namespace string, selector labels.Selector) (ret []*Volume, err error)
	Get(namespace, name string) (*Volume, error)
}

type VolumeController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() VolumeLister
	AddHandler(ctx context.Context, name string, handler VolumeHandlerFunc)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, handler VolumeHandlerFunc)
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
	AddHandler(ctx context.Context, name string, sync VolumeHandlerFunc)
	AddLifecycle(ctx context.Context, name string, lifecycle VolumeLifecycle)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync VolumeHandlerFunc)
	AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle VolumeLifecycle)
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

func (c *volumeController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *volumeController) Lister() VolumeLister {
	return &volumeLister{
		controller: c,
	}
}

func (c *volumeController) AddHandler(ctx context.Context, name string, handler VolumeHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*Volume); ok {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

func (c *volumeController) AddClusterScopedHandler(ctx context.Context, name, cluster string, handler VolumeHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*Volume); ok && controller.ObjectInCluster(cluster, obj) {
			return handler(key, v)
		} else {
			return nil, nil
		}
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
func (s *volumeClient) Patch(o *Volume, patchType types.PatchType, data []byte, subresources ...string) (*Volume, error) {
	obj, err := s.objectClient.Patch(o.Name, o, patchType, data, subresources...)
	return obj.(*Volume), err
}

func (s *volumeClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *volumeClient) AddHandler(ctx context.Context, name string, sync VolumeHandlerFunc) {
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *volumeClient) AddLifecycle(ctx context.Context, name string, lifecycle VolumeLifecycle) {
	sync := NewVolumeLifecycleAdapter(name, false, s, lifecycle)
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *volumeClient) AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync VolumeHandlerFunc) {
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

func (s *volumeClient) AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle VolumeLifecycle) {
	sync := NewVolumeLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

type VolumeIndexer func(obj *Volume) ([]string, error)

type VolumeClientCache interface {
	Get(namespace, name string) (*Volume, error)
	List(namespace string, selector labels.Selector) ([]*Volume, error)

	Index(name string, indexer VolumeIndexer)
	GetIndexed(name, key string) ([]*Volume, error)
}

type VolumeClient interface {
	Create(*Volume) (*Volume, error)
	Get(namespace, name string, opts metav1.GetOptions) (*Volume, error)
	Update(*Volume) (*Volume, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	List(namespace string, opts metav1.ListOptions) (*VolumeList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)

	Cache() VolumeClientCache

	OnCreate(ctx context.Context, name string, sync VolumeChangeHandlerFunc)
	OnChange(ctx context.Context, name string, sync VolumeChangeHandlerFunc)
	OnRemove(ctx context.Context, name string, sync VolumeChangeHandlerFunc)
	Enqueue(namespace, name string)

	Generic() controller.GenericController
	ObjectClient() *objectclient.ObjectClient
	Interface() VolumeInterface
}

type volumeClientCache struct {
	client *volumeClient2
}

type volumeClient2 struct {
	iface      VolumeInterface
	controller VolumeController
}

func (n *volumeClient2) Interface() VolumeInterface {
	return n.iface
}

func (n *volumeClient2) Generic() controller.GenericController {
	return n.iface.Controller().Generic()
}

func (n *volumeClient2) ObjectClient() *objectclient.ObjectClient {
	return n.Interface().ObjectClient()
}

func (n *volumeClient2) Enqueue(namespace, name string) {
	n.iface.Controller().Enqueue(namespace, name)
}

func (n *volumeClient2) Create(obj *Volume) (*Volume, error) {
	return n.iface.Create(obj)
}

func (n *volumeClient2) Get(namespace, name string, opts metav1.GetOptions) (*Volume, error) {
	return n.iface.GetNamespaced(namespace, name, opts)
}

func (n *volumeClient2) Update(obj *Volume) (*Volume, error) {
	return n.iface.Update(obj)
}

func (n *volumeClient2) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	return n.iface.DeleteNamespaced(namespace, name, options)
}

func (n *volumeClient2) List(namespace string, opts metav1.ListOptions) (*VolumeList, error) {
	return n.iface.List(opts)
}

func (n *volumeClient2) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return n.iface.Watch(opts)
}

func (n *volumeClientCache) Get(namespace, name string) (*Volume, error) {
	return n.client.controller.Lister().Get(namespace, name)
}

func (n *volumeClientCache) List(namespace string, selector labels.Selector) ([]*Volume, error) {
	return n.client.controller.Lister().List(namespace, selector)
}

func (n *volumeClient2) Cache() VolumeClientCache {
	n.loadController()
	return &volumeClientCache{
		client: n,
	}
}

func (n *volumeClient2) OnCreate(ctx context.Context, name string, sync VolumeChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-create", &volumeLifecycleDelegate{create: sync})
}

func (n *volumeClient2) OnChange(ctx context.Context, name string, sync VolumeChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-change", &volumeLifecycleDelegate{update: sync})
}

func (n *volumeClient2) OnRemove(ctx context.Context, name string, sync VolumeChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name, &volumeLifecycleDelegate{remove: sync})
}

func (n *volumeClientCache) Index(name string, indexer VolumeIndexer) {
	err := n.client.controller.Informer().GetIndexer().AddIndexers(map[string]cache.IndexFunc{
		name: func(obj interface{}) ([]string, error) {
			if v, ok := obj.(*Volume); ok {
				return indexer(v)
			}
			return nil, nil
		},
	})

	if err != nil {
		panic(err)
	}
}

func (n *volumeClientCache) GetIndexed(name, key string) ([]*Volume, error) {
	var result []*Volume
	objs, err := n.client.controller.Informer().GetIndexer().ByIndex(name, key)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		if v, ok := obj.(*Volume); ok {
			result = append(result, v)
		}
	}

	return result, nil
}

func (n *volumeClient2) loadController() {
	if n.controller == nil {
		n.controller = n.iface.Controller()
	}
}

type volumeLifecycleDelegate struct {
	create VolumeChangeHandlerFunc
	update VolumeChangeHandlerFunc
	remove VolumeChangeHandlerFunc
}

func (n *volumeLifecycleDelegate) HasCreate() bool {
	return n.create != nil
}

func (n *volumeLifecycleDelegate) Create(obj *Volume) (runtime.Object, error) {
	if n.create == nil {
		return obj, nil
	}
	return n.create(obj)
}

func (n *volumeLifecycleDelegate) HasFinalize() bool {
	return n.remove != nil
}

func (n *volumeLifecycleDelegate) Remove(obj *Volume) (runtime.Object, error) {
	if n.remove == nil {
		return obj, nil
	}
	return n.remove(obj)
}

func (n *volumeLifecycleDelegate) Updated(obj *Volume) (runtime.Object, error) {
	if n.update == nil {
		return obj, nil
	}
	return n.update(obj)
}
