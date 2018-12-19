package v1beta1

import (
	"context"

	"github.com/rancher/norman/controller"
	"github.com/rancher/norman/objectclient"
	"k8s.io/api/policy/v1beta1"
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
	PodDisruptionBudgetGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "PodDisruptionBudget",
	}
	PodDisruptionBudgetResource = metav1.APIResource{
		Name:         "poddisruptionbudgets",
		SingularName: "poddisruptionbudget",
		Namespaced:   true,

		Kind: PodDisruptionBudgetGroupVersionKind.Kind,
	}
)

func NewPodDisruptionBudget(namespace, name string, obj v1beta1.PodDisruptionBudget) *v1beta1.PodDisruptionBudget {
	obj.APIVersion, obj.Kind = PodDisruptionBudgetGroupVersionKind.ToAPIVersionAndKind()
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

type PodDisruptionBudgetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []v1beta1.PodDisruptionBudget
}

type PodDisruptionBudgetHandlerFunc func(key string, obj *v1beta1.PodDisruptionBudget) (runtime.Object, error)

type PodDisruptionBudgetChangeHandlerFunc func(obj *v1beta1.PodDisruptionBudget) (runtime.Object, error)

type PodDisruptionBudgetLister interface {
	List(namespace string, selector labels.Selector) (ret []*v1beta1.PodDisruptionBudget, err error)
	Get(namespace, name string) (*v1beta1.PodDisruptionBudget, error)
}

type PodDisruptionBudgetController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() PodDisruptionBudgetLister
	AddHandler(ctx context.Context, name string, handler PodDisruptionBudgetHandlerFunc)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, handler PodDisruptionBudgetHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type PodDisruptionBudgetInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*v1beta1.PodDisruptionBudget) (*v1beta1.PodDisruptionBudget, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*v1beta1.PodDisruptionBudget, error)
	Get(name string, opts metav1.GetOptions) (*v1beta1.PodDisruptionBudget, error)
	Update(*v1beta1.PodDisruptionBudget) (*v1beta1.PodDisruptionBudget, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*PodDisruptionBudgetList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() PodDisruptionBudgetController
	AddHandler(ctx context.Context, name string, sync PodDisruptionBudgetHandlerFunc)
	AddLifecycle(ctx context.Context, name string, lifecycle PodDisruptionBudgetLifecycle)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync PodDisruptionBudgetHandlerFunc)
	AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle PodDisruptionBudgetLifecycle)
}

type podDisruptionBudgetLister struct {
	controller *podDisruptionBudgetController
}

func (l *podDisruptionBudgetLister) List(namespace string, selector labels.Selector) (ret []*v1beta1.PodDisruptionBudget, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*v1beta1.PodDisruptionBudget))
	})
	return
}

func (l *podDisruptionBudgetLister) Get(namespace, name string) (*v1beta1.PodDisruptionBudget, error) {
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
			Group:    PodDisruptionBudgetGroupVersionKind.Group,
			Resource: "podDisruptionBudget",
		}, key)
	}
	return obj.(*v1beta1.PodDisruptionBudget), nil
}

type podDisruptionBudgetController struct {
	controller.GenericController
}

func (c *podDisruptionBudgetController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *podDisruptionBudgetController) Lister() PodDisruptionBudgetLister {
	return &podDisruptionBudgetLister{
		controller: c,
	}
}

func (c *podDisruptionBudgetController) AddHandler(ctx context.Context, name string, handler PodDisruptionBudgetHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*v1beta1.PodDisruptionBudget); ok {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

func (c *podDisruptionBudgetController) AddClusterScopedHandler(ctx context.Context, name, cluster string, handler PodDisruptionBudgetHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*v1beta1.PodDisruptionBudget); ok && controller.ObjectInCluster(cluster, obj) {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

type podDisruptionBudgetFactory struct {
}

func (c podDisruptionBudgetFactory) Object() runtime.Object {
	return &v1beta1.PodDisruptionBudget{}
}

func (c podDisruptionBudgetFactory) List() runtime.Object {
	return &PodDisruptionBudgetList{}
}

func (s *podDisruptionBudgetClient) Controller() PodDisruptionBudgetController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.podDisruptionBudgetControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(PodDisruptionBudgetGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &podDisruptionBudgetController{
		GenericController: genericController,
	}

	s.client.podDisruptionBudgetControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type podDisruptionBudgetClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   PodDisruptionBudgetController
}

func (s *podDisruptionBudgetClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *podDisruptionBudgetClient) Create(o *v1beta1.PodDisruptionBudget) (*v1beta1.PodDisruptionBudget, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*v1beta1.PodDisruptionBudget), err
}

func (s *podDisruptionBudgetClient) Get(name string, opts metav1.GetOptions) (*v1beta1.PodDisruptionBudget, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*v1beta1.PodDisruptionBudget), err
}

func (s *podDisruptionBudgetClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*v1beta1.PodDisruptionBudget, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*v1beta1.PodDisruptionBudget), err
}

func (s *podDisruptionBudgetClient) Update(o *v1beta1.PodDisruptionBudget) (*v1beta1.PodDisruptionBudget, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*v1beta1.PodDisruptionBudget), err
}

func (s *podDisruptionBudgetClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *podDisruptionBudgetClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *podDisruptionBudgetClient) List(opts metav1.ListOptions) (*PodDisruptionBudgetList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*PodDisruptionBudgetList), err
}

func (s *podDisruptionBudgetClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *podDisruptionBudgetClient) Patch(o *v1beta1.PodDisruptionBudget, patchType types.PatchType, data []byte, subresources ...string) (*v1beta1.PodDisruptionBudget, error) {
	obj, err := s.objectClient.Patch(o.Name, o, patchType, data, subresources...)
	return obj.(*v1beta1.PodDisruptionBudget), err
}

func (s *podDisruptionBudgetClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *podDisruptionBudgetClient) AddHandler(ctx context.Context, name string, sync PodDisruptionBudgetHandlerFunc) {
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *podDisruptionBudgetClient) AddLifecycle(ctx context.Context, name string, lifecycle PodDisruptionBudgetLifecycle) {
	sync := NewPodDisruptionBudgetLifecycleAdapter(name, false, s, lifecycle)
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *podDisruptionBudgetClient) AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync PodDisruptionBudgetHandlerFunc) {
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

func (s *podDisruptionBudgetClient) AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle PodDisruptionBudgetLifecycle) {
	sync := NewPodDisruptionBudgetLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

type PodDisruptionBudgetIndexer func(obj *v1beta1.PodDisruptionBudget) ([]string, error)

type PodDisruptionBudgetClientCache interface {
	Get(namespace, name string) (*v1beta1.PodDisruptionBudget, error)
	List(namespace string, selector labels.Selector) ([]*v1beta1.PodDisruptionBudget, error)

	Index(name string, indexer PodDisruptionBudgetIndexer)
	GetIndexed(name, key string) ([]*v1beta1.PodDisruptionBudget, error)
}

type PodDisruptionBudgetClient interface {
	Create(*v1beta1.PodDisruptionBudget) (*v1beta1.PodDisruptionBudget, error)
	Get(namespace, name string, opts metav1.GetOptions) (*v1beta1.PodDisruptionBudget, error)
	Update(*v1beta1.PodDisruptionBudget) (*v1beta1.PodDisruptionBudget, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	List(namespace string, opts metav1.ListOptions) (*PodDisruptionBudgetList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)

	Cache() PodDisruptionBudgetClientCache

	OnCreate(ctx context.Context, name string, sync PodDisruptionBudgetChangeHandlerFunc)
	OnChange(ctx context.Context, name string, sync PodDisruptionBudgetChangeHandlerFunc)
	OnRemove(ctx context.Context, name string, sync PodDisruptionBudgetChangeHandlerFunc)
	Enqueue(namespace, name string)

	Generic() controller.GenericController
	ObjectClient() *objectclient.ObjectClient
	Interface() PodDisruptionBudgetInterface
}

type podDisruptionBudgetClientCache struct {
	client *podDisruptionBudgetClient2
}

type podDisruptionBudgetClient2 struct {
	iface      PodDisruptionBudgetInterface
	controller PodDisruptionBudgetController
}

func (n *podDisruptionBudgetClient2) Interface() PodDisruptionBudgetInterface {
	return n.iface
}

func (n *podDisruptionBudgetClient2) Generic() controller.GenericController {
	return n.iface.Controller().Generic()
}

func (n *podDisruptionBudgetClient2) ObjectClient() *objectclient.ObjectClient {
	return n.Interface().ObjectClient()
}

func (n *podDisruptionBudgetClient2) Enqueue(namespace, name string) {
	n.iface.Controller().Enqueue(namespace, name)
}

func (n *podDisruptionBudgetClient2) Create(obj *v1beta1.PodDisruptionBudget) (*v1beta1.PodDisruptionBudget, error) {
	return n.iface.Create(obj)
}

func (n *podDisruptionBudgetClient2) Get(namespace, name string, opts metav1.GetOptions) (*v1beta1.PodDisruptionBudget, error) {
	return n.iface.GetNamespaced(namespace, name, opts)
}

func (n *podDisruptionBudgetClient2) Update(obj *v1beta1.PodDisruptionBudget) (*v1beta1.PodDisruptionBudget, error) {
	return n.iface.Update(obj)
}

func (n *podDisruptionBudgetClient2) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	return n.iface.DeleteNamespaced(namespace, name, options)
}

func (n *podDisruptionBudgetClient2) List(namespace string, opts metav1.ListOptions) (*PodDisruptionBudgetList, error) {
	return n.iface.List(opts)
}

func (n *podDisruptionBudgetClient2) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return n.iface.Watch(opts)
}

func (n *podDisruptionBudgetClientCache) Get(namespace, name string) (*v1beta1.PodDisruptionBudget, error) {
	return n.client.controller.Lister().Get(namespace, name)
}

func (n *podDisruptionBudgetClientCache) List(namespace string, selector labels.Selector) ([]*v1beta1.PodDisruptionBudget, error) {
	return n.client.controller.Lister().List(namespace, selector)
}

func (n *podDisruptionBudgetClient2) Cache() PodDisruptionBudgetClientCache {
	n.loadController()
	return &podDisruptionBudgetClientCache{
		client: n,
	}
}

func (n *podDisruptionBudgetClient2) OnCreate(ctx context.Context, name string, sync PodDisruptionBudgetChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-create", &podDisruptionBudgetLifecycleDelegate{create: sync})
}

func (n *podDisruptionBudgetClient2) OnChange(ctx context.Context, name string, sync PodDisruptionBudgetChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-change", &podDisruptionBudgetLifecycleDelegate{update: sync})
}

func (n *podDisruptionBudgetClient2) OnRemove(ctx context.Context, name string, sync PodDisruptionBudgetChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name, &podDisruptionBudgetLifecycleDelegate{remove: sync})
}

func (n *podDisruptionBudgetClientCache) Index(name string, indexer PodDisruptionBudgetIndexer) {
	err := n.client.controller.Informer().GetIndexer().AddIndexers(map[string]cache.IndexFunc{
		name: func(obj interface{}) ([]string, error) {
			if v, ok := obj.(*v1beta1.PodDisruptionBudget); ok {
				return indexer(v)
			}
			return nil, nil
		},
	})

	if err != nil {
		panic(err)
	}
}

func (n *podDisruptionBudgetClientCache) GetIndexed(name, key string) ([]*v1beta1.PodDisruptionBudget, error) {
	var result []*v1beta1.PodDisruptionBudget
	objs, err := n.client.controller.Informer().GetIndexer().ByIndex(name, key)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		if v, ok := obj.(*v1beta1.PodDisruptionBudget); ok {
			result = append(result, v)
		}
	}

	return result, nil
}

func (n *podDisruptionBudgetClient2) loadController() {
	if n.controller == nil {
		n.controller = n.iface.Controller()
	}
}

type podDisruptionBudgetLifecycleDelegate struct {
	create PodDisruptionBudgetChangeHandlerFunc
	update PodDisruptionBudgetChangeHandlerFunc
	remove PodDisruptionBudgetChangeHandlerFunc
}

func (n *podDisruptionBudgetLifecycleDelegate) HasCreate() bool {
	return n.create != nil
}

func (n *podDisruptionBudgetLifecycleDelegate) Create(obj *v1beta1.PodDisruptionBudget) (runtime.Object, error) {
	if n.create == nil {
		return obj, nil
	}
	return n.create(obj)
}

func (n *podDisruptionBudgetLifecycleDelegate) HasFinalize() bool {
	return n.remove != nil
}

func (n *podDisruptionBudgetLifecycleDelegate) Remove(obj *v1beta1.PodDisruptionBudget) (runtime.Object, error) {
	if n.remove == nil {
		return obj, nil
	}
	return n.remove(obj)
}

func (n *podDisruptionBudgetLifecycleDelegate) Updated(obj *v1beta1.PodDisruptionBudget) (runtime.Object, error) {
	if n.update == nil {
		return obj, nil
	}
	return n.update(obj)
}
