package v1alpha1

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
	BuildGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "Build",
	}
	BuildResource = metav1.APIResource{
		Name:         "builds",
		SingularName: "build",
		Namespaced:   true,

		Kind: BuildGroupVersionKind.Kind,
	}
)

func NewBuild(namespace, name string, obj Build) *Build {
	obj.APIVersion, obj.Kind = BuildGroupVersionKind.ToAPIVersionAndKind()
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

type BuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Build
}

type BuildHandlerFunc func(key string, obj *Build) (runtime.Object, error)

type BuildChangeHandlerFunc func(obj *Build) (runtime.Object, error)

type BuildLister interface {
	List(namespace string, selector labels.Selector) (ret []*Build, err error)
	Get(namespace, name string) (*Build, error)
}

type BuildController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() BuildLister
	AddHandler(ctx context.Context, name string, handler BuildHandlerFunc)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, handler BuildHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type BuildInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*Build) (*Build, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*Build, error)
	Get(name string, opts metav1.GetOptions) (*Build, error)
	Update(*Build) (*Build, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*BuildList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() BuildController
	AddHandler(ctx context.Context, name string, sync BuildHandlerFunc)
	AddLifecycle(ctx context.Context, name string, lifecycle BuildLifecycle)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync BuildHandlerFunc)
	AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle BuildLifecycle)
}

type buildLister struct {
	controller *buildController
}

func (l *buildLister) List(namespace string, selector labels.Selector) (ret []*Build, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*Build))
	})
	return
}

func (l *buildLister) Get(namespace, name string) (*Build, error) {
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
			Group:    BuildGroupVersionKind.Group,
			Resource: "build",
		}, key)
	}
	return obj.(*Build), nil
}

type buildController struct {
	controller.GenericController
}

func (c *buildController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *buildController) Lister() BuildLister {
	return &buildLister{
		controller: c,
	}
}

func (c *buildController) AddHandler(ctx context.Context, name string, handler BuildHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*Build); ok {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

func (c *buildController) AddClusterScopedHandler(ctx context.Context, name, cluster string, handler BuildHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*Build); ok && controller.ObjectInCluster(cluster, obj) {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

type buildFactory struct {
}

func (c buildFactory) Object() runtime.Object {
	return &Build{}
}

func (c buildFactory) List() runtime.Object {
	return &BuildList{}
}

func (s *buildClient) Controller() BuildController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.buildControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(BuildGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &buildController{
		GenericController: genericController,
	}

	s.client.buildControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type buildClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   BuildController
}

func (s *buildClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *buildClient) Create(o *Build) (*Build, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*Build), err
}

func (s *buildClient) Get(name string, opts metav1.GetOptions) (*Build, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*Build), err
}

func (s *buildClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*Build, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*Build), err
}

func (s *buildClient) Update(o *Build) (*Build, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*Build), err
}

func (s *buildClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *buildClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *buildClient) List(opts metav1.ListOptions) (*BuildList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*BuildList), err
}

func (s *buildClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *buildClient) Patch(o *Build, patchType types.PatchType, data []byte, subresources ...string) (*Build, error) {
	obj, err := s.objectClient.Patch(o.Name, o, patchType, data, subresources...)
	return obj.(*Build), err
}

func (s *buildClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *buildClient) AddHandler(ctx context.Context, name string, sync BuildHandlerFunc) {
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *buildClient) AddLifecycle(ctx context.Context, name string, lifecycle BuildLifecycle) {
	sync := NewBuildLifecycleAdapter(name, false, s, lifecycle)
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *buildClient) AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync BuildHandlerFunc) {
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

func (s *buildClient) AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle BuildLifecycle) {
	sync := NewBuildLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

type BuildIndexer func(obj *Build) ([]string, error)

type BuildClientCache interface {
	Get(namespace, name string) (*Build, error)
	List(namespace string, selector labels.Selector) ([]*Build, error)

	Index(name string, indexer BuildIndexer)
	GetIndexed(name, key string) ([]*Build, error)
}

type BuildClient interface {
	Create(*Build) (*Build, error)
	Get(namespace, name string, opts metav1.GetOptions) (*Build, error)
	Update(*Build) (*Build, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	List(namespace string, opts metav1.ListOptions) (*BuildList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)

	Cache() BuildClientCache

	OnCreate(ctx context.Context, name string, sync BuildChangeHandlerFunc)
	OnChange(ctx context.Context, name string, sync BuildChangeHandlerFunc)
	OnRemove(ctx context.Context, name string, sync BuildChangeHandlerFunc)
	Enqueue(namespace, name string)

	Generic() controller.GenericController
	ObjectClient() *objectclient.ObjectClient
	Interface() BuildInterface
}

type buildClientCache struct {
	client *buildClient2
}

type buildClient2 struct {
	iface      BuildInterface
	controller BuildController
}

func (n *buildClient2) Interface() BuildInterface {
	return n.iface
}

func (n *buildClient2) Generic() controller.GenericController {
	return n.iface.Controller().Generic()
}

func (n *buildClient2) ObjectClient() *objectclient.ObjectClient {
	return n.Interface().ObjectClient()
}

func (n *buildClient2) Enqueue(namespace, name string) {
	n.iface.Controller().Enqueue(namespace, name)
}

func (n *buildClient2) Create(obj *Build) (*Build, error) {
	return n.iface.Create(obj)
}

func (n *buildClient2) Get(namespace, name string, opts metav1.GetOptions) (*Build, error) {
	return n.iface.GetNamespaced(namespace, name, opts)
}

func (n *buildClient2) Update(obj *Build) (*Build, error) {
	return n.iface.Update(obj)
}

func (n *buildClient2) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	return n.iface.DeleteNamespaced(namespace, name, options)
}

func (n *buildClient2) List(namespace string, opts metav1.ListOptions) (*BuildList, error) {
	return n.iface.List(opts)
}

func (n *buildClient2) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return n.iface.Watch(opts)
}

func (n *buildClientCache) Get(namespace, name string) (*Build, error) {
	return n.client.controller.Lister().Get(namespace, name)
}

func (n *buildClientCache) List(namespace string, selector labels.Selector) ([]*Build, error) {
	return n.client.controller.Lister().List(namespace, selector)
}

func (n *buildClient2) Cache() BuildClientCache {
	n.loadController()
	return &buildClientCache{
		client: n,
	}
}

func (n *buildClient2) OnCreate(ctx context.Context, name string, sync BuildChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-create", &buildLifecycleDelegate{create: sync})
}

func (n *buildClient2) OnChange(ctx context.Context, name string, sync BuildChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-change", &buildLifecycleDelegate{update: sync})
}

func (n *buildClient2) OnRemove(ctx context.Context, name string, sync BuildChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name, &buildLifecycleDelegate{remove: sync})
}

func (n *buildClientCache) Index(name string, indexer BuildIndexer) {
	err := n.client.controller.Informer().GetIndexer().AddIndexers(map[string]cache.IndexFunc{
		name: func(obj interface{}) ([]string, error) {
			if v, ok := obj.(*Build); ok {
				return indexer(v)
			}
			return nil, nil
		},
	})

	if err != nil {
		panic(err)
	}
}

func (n *buildClientCache) GetIndexed(name, key string) ([]*Build, error) {
	var result []*Build
	objs, err := n.client.controller.Informer().GetIndexer().ByIndex(name, key)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		if v, ok := obj.(*Build); ok {
			result = append(result, v)
		}
	}

	return result, nil
}

func (n *buildClient2) loadController() {
	if n.controller == nil {
		n.controller = n.iface.Controller()
	}
}

type buildLifecycleDelegate struct {
	create BuildChangeHandlerFunc
	update BuildChangeHandlerFunc
	remove BuildChangeHandlerFunc
}

func (n *buildLifecycleDelegate) HasCreate() bool {
	return n.create != nil
}

func (n *buildLifecycleDelegate) Create(obj *Build) (runtime.Object, error) {
	if n.create == nil {
		return obj, nil
	}
	return n.create(obj)
}

func (n *buildLifecycleDelegate) HasFinalize() bool {
	return n.remove != nil
}

func (n *buildLifecycleDelegate) Remove(obj *Build) (runtime.Object, error) {
	if n.remove == nil {
		return obj, nil
	}
	return n.remove(obj)
}

func (n *buildLifecycleDelegate) Updated(obj *Build) (runtime.Object, error) {
	if n.update == nil {
		return obj, nil
	}
	return n.update(obj)
}
