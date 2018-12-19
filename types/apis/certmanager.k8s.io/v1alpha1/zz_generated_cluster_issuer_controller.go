package v1alpha1

import (
	"context"

	"github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
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
	ClusterIssuerGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "ClusterIssuer",
	}
	ClusterIssuerResource = metav1.APIResource{
		Name:         "clusterissuers",
		SingularName: "clusterissuer",
		Namespaced:   false,
		Kind:         ClusterIssuerGroupVersionKind.Kind,
	}
)

func NewClusterIssuer(namespace, name string, obj v1alpha1.ClusterIssuer) *v1alpha1.ClusterIssuer {
	obj.APIVersion, obj.Kind = ClusterIssuerGroupVersionKind.ToAPIVersionAndKind()
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

type ClusterIssuerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []v1alpha1.ClusterIssuer
}

type ClusterIssuerHandlerFunc func(key string, obj *v1alpha1.ClusterIssuer) (runtime.Object, error)

type ClusterIssuerChangeHandlerFunc func(obj *v1alpha1.ClusterIssuer) (runtime.Object, error)

type ClusterIssuerLister interface {
	List(namespace string, selector labels.Selector) (ret []*v1alpha1.ClusterIssuer, err error)
	Get(namespace, name string) (*v1alpha1.ClusterIssuer, error)
}

type ClusterIssuerController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() ClusterIssuerLister
	AddHandler(ctx context.Context, name string, handler ClusterIssuerHandlerFunc)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, handler ClusterIssuerHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type ClusterIssuerInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*v1alpha1.ClusterIssuer) (*v1alpha1.ClusterIssuer, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*v1alpha1.ClusterIssuer, error)
	Get(name string, opts metav1.GetOptions) (*v1alpha1.ClusterIssuer, error)
	Update(*v1alpha1.ClusterIssuer) (*v1alpha1.ClusterIssuer, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*ClusterIssuerList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() ClusterIssuerController
	AddHandler(ctx context.Context, name string, sync ClusterIssuerHandlerFunc)
	AddLifecycle(ctx context.Context, name string, lifecycle ClusterIssuerLifecycle)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync ClusterIssuerHandlerFunc)
	AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle ClusterIssuerLifecycle)
}

type clusterIssuerLister struct {
	controller *clusterIssuerController
}

func (l *clusterIssuerLister) List(namespace string, selector labels.Selector) (ret []*v1alpha1.ClusterIssuer, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*v1alpha1.ClusterIssuer))
	})
	return
}

func (l *clusterIssuerLister) Get(namespace, name string) (*v1alpha1.ClusterIssuer, error) {
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
			Group:    ClusterIssuerGroupVersionKind.Group,
			Resource: "clusterIssuer",
		}, key)
	}
	return obj.(*v1alpha1.ClusterIssuer), nil
}

type clusterIssuerController struct {
	controller.GenericController
}

func (c *clusterIssuerController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *clusterIssuerController) Lister() ClusterIssuerLister {
	return &clusterIssuerLister{
		controller: c,
	}
}

func (c *clusterIssuerController) AddHandler(ctx context.Context, name string, handler ClusterIssuerHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*v1alpha1.ClusterIssuer); ok {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

func (c *clusterIssuerController) AddClusterScopedHandler(ctx context.Context, name, cluster string, handler ClusterIssuerHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*v1alpha1.ClusterIssuer); ok && controller.ObjectInCluster(cluster, obj) {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

type clusterIssuerFactory struct {
}

func (c clusterIssuerFactory) Object() runtime.Object {
	return &v1alpha1.ClusterIssuer{}
}

func (c clusterIssuerFactory) List() runtime.Object {
	return &ClusterIssuerList{}
}

func (s *clusterIssuerClient) Controller() ClusterIssuerController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.clusterIssuerControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(ClusterIssuerGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &clusterIssuerController{
		GenericController: genericController,
	}

	s.client.clusterIssuerControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type clusterIssuerClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   ClusterIssuerController
}

func (s *clusterIssuerClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *clusterIssuerClient) Create(o *v1alpha1.ClusterIssuer) (*v1alpha1.ClusterIssuer, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*v1alpha1.ClusterIssuer), err
}

func (s *clusterIssuerClient) Get(name string, opts metav1.GetOptions) (*v1alpha1.ClusterIssuer, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*v1alpha1.ClusterIssuer), err
}

func (s *clusterIssuerClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*v1alpha1.ClusterIssuer, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*v1alpha1.ClusterIssuer), err
}

func (s *clusterIssuerClient) Update(o *v1alpha1.ClusterIssuer) (*v1alpha1.ClusterIssuer, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*v1alpha1.ClusterIssuer), err
}

func (s *clusterIssuerClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *clusterIssuerClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *clusterIssuerClient) List(opts metav1.ListOptions) (*ClusterIssuerList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*ClusterIssuerList), err
}

func (s *clusterIssuerClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *clusterIssuerClient) Patch(o *v1alpha1.ClusterIssuer, patchType types.PatchType, data []byte, subresources ...string) (*v1alpha1.ClusterIssuer, error) {
	obj, err := s.objectClient.Patch(o.Name, o, patchType, data, subresources...)
	return obj.(*v1alpha1.ClusterIssuer), err
}

func (s *clusterIssuerClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *clusterIssuerClient) AddHandler(ctx context.Context, name string, sync ClusterIssuerHandlerFunc) {
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *clusterIssuerClient) AddLifecycle(ctx context.Context, name string, lifecycle ClusterIssuerLifecycle) {
	sync := NewClusterIssuerLifecycleAdapter(name, false, s, lifecycle)
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *clusterIssuerClient) AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync ClusterIssuerHandlerFunc) {
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

func (s *clusterIssuerClient) AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle ClusterIssuerLifecycle) {
	sync := NewClusterIssuerLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

type ClusterIssuerIndexer func(obj *v1alpha1.ClusterIssuer) ([]string, error)

type ClusterIssuerClientCache interface {
	Get(namespace, name string) (*v1alpha1.ClusterIssuer, error)
	List(namespace string, selector labels.Selector) ([]*v1alpha1.ClusterIssuer, error)

	Index(name string, indexer ClusterIssuerIndexer)
	GetIndexed(name, key string) ([]*v1alpha1.ClusterIssuer, error)
}

type ClusterIssuerClient interface {
	Create(*v1alpha1.ClusterIssuer) (*v1alpha1.ClusterIssuer, error)
	Get(namespace, name string, opts metav1.GetOptions) (*v1alpha1.ClusterIssuer, error)
	Update(*v1alpha1.ClusterIssuer) (*v1alpha1.ClusterIssuer, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	List(namespace string, opts metav1.ListOptions) (*ClusterIssuerList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)

	Cache() ClusterIssuerClientCache

	OnCreate(ctx context.Context, name string, sync ClusterIssuerChangeHandlerFunc)
	OnChange(ctx context.Context, name string, sync ClusterIssuerChangeHandlerFunc)
	OnRemove(ctx context.Context, name string, sync ClusterIssuerChangeHandlerFunc)
	Enqueue(namespace, name string)

	Generic() controller.GenericController
	ObjectClient() *objectclient.ObjectClient
	Interface() ClusterIssuerInterface
}

type clusterIssuerClientCache struct {
	client *clusterIssuerClient2
}

type clusterIssuerClient2 struct {
	iface      ClusterIssuerInterface
	controller ClusterIssuerController
}

func (n *clusterIssuerClient2) Interface() ClusterIssuerInterface {
	return n.iface
}

func (n *clusterIssuerClient2) Generic() controller.GenericController {
	return n.iface.Controller().Generic()
}

func (n *clusterIssuerClient2) ObjectClient() *objectclient.ObjectClient {
	return n.Interface().ObjectClient()
}

func (n *clusterIssuerClient2) Enqueue(namespace, name string) {
	n.iface.Controller().Enqueue(namespace, name)
}

func (n *clusterIssuerClient2) Create(obj *v1alpha1.ClusterIssuer) (*v1alpha1.ClusterIssuer, error) {
	return n.iface.Create(obj)
}

func (n *clusterIssuerClient2) Get(namespace, name string, opts metav1.GetOptions) (*v1alpha1.ClusterIssuer, error) {
	return n.iface.GetNamespaced(namespace, name, opts)
}

func (n *clusterIssuerClient2) Update(obj *v1alpha1.ClusterIssuer) (*v1alpha1.ClusterIssuer, error) {
	return n.iface.Update(obj)
}

func (n *clusterIssuerClient2) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	return n.iface.DeleteNamespaced(namespace, name, options)
}

func (n *clusterIssuerClient2) List(namespace string, opts metav1.ListOptions) (*ClusterIssuerList, error) {
	return n.iface.List(opts)
}

func (n *clusterIssuerClient2) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return n.iface.Watch(opts)
}

func (n *clusterIssuerClientCache) Get(namespace, name string) (*v1alpha1.ClusterIssuer, error) {
	return n.client.controller.Lister().Get(namespace, name)
}

func (n *clusterIssuerClientCache) List(namespace string, selector labels.Selector) ([]*v1alpha1.ClusterIssuer, error) {
	return n.client.controller.Lister().List(namespace, selector)
}

func (n *clusterIssuerClient2) Cache() ClusterIssuerClientCache {
	n.loadController()
	return &clusterIssuerClientCache{
		client: n,
	}
}

func (n *clusterIssuerClient2) OnCreate(ctx context.Context, name string, sync ClusterIssuerChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-create", &clusterIssuerLifecycleDelegate{create: sync})
}

func (n *clusterIssuerClient2) OnChange(ctx context.Context, name string, sync ClusterIssuerChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-change", &clusterIssuerLifecycleDelegate{update: sync})
}

func (n *clusterIssuerClient2) OnRemove(ctx context.Context, name string, sync ClusterIssuerChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name, &clusterIssuerLifecycleDelegate{remove: sync})
}

func (n *clusterIssuerClientCache) Index(name string, indexer ClusterIssuerIndexer) {
	err := n.client.controller.Informer().GetIndexer().AddIndexers(map[string]cache.IndexFunc{
		name: func(obj interface{}) ([]string, error) {
			if v, ok := obj.(*v1alpha1.ClusterIssuer); ok {
				return indexer(v)
			}
			return nil, nil
		},
	})

	if err != nil {
		panic(err)
	}
}

func (n *clusterIssuerClientCache) GetIndexed(name, key string) ([]*v1alpha1.ClusterIssuer, error) {
	var result []*v1alpha1.ClusterIssuer
	objs, err := n.client.controller.Informer().GetIndexer().ByIndex(name, key)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		if v, ok := obj.(*v1alpha1.ClusterIssuer); ok {
			result = append(result, v)
		}
	}

	return result, nil
}

func (n *clusterIssuerClient2) loadController() {
	if n.controller == nil {
		n.controller = n.iface.Controller()
	}
}

type clusterIssuerLifecycleDelegate struct {
	create ClusterIssuerChangeHandlerFunc
	update ClusterIssuerChangeHandlerFunc
	remove ClusterIssuerChangeHandlerFunc
}

func (n *clusterIssuerLifecycleDelegate) HasCreate() bool {
	return n.create != nil
}

func (n *clusterIssuerLifecycleDelegate) Create(obj *v1alpha1.ClusterIssuer) (runtime.Object, error) {
	if n.create == nil {
		return obj, nil
	}
	return n.create(obj)
}

func (n *clusterIssuerLifecycleDelegate) HasFinalize() bool {
	return n.remove != nil
}

func (n *clusterIssuerLifecycleDelegate) Remove(obj *v1alpha1.ClusterIssuer) (runtime.Object, error) {
	if n.remove == nil {
		return obj, nil
	}
	return n.remove(obj)
}

func (n *clusterIssuerLifecycleDelegate) Updated(obj *v1alpha1.ClusterIssuer) (runtime.Object, error) {
	if n.update == nil {
		return obj, nil
	}
	return n.update(obj)
}
