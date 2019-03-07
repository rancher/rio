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
	GitWebHookExecutionGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "GitWebHookExecution",
	}
	GitWebHookExecutionResource = metav1.APIResource{
		Name:         "gitwebhookexecutions",
		SingularName: "gitwebhookexecution",
		Namespaced:   true,

		Kind: GitWebHookExecutionGroupVersionKind.Kind,
	}
)

func NewGitWebHookExecution(namespace, name string, obj GitWebHookExecution) *GitWebHookExecution {
	obj.APIVersion, obj.Kind = GitWebHookExecutionGroupVersionKind.ToAPIVersionAndKind()
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

type GitWebHookExecutionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitWebHookExecution
}

type GitWebHookExecutionHandlerFunc func(key string, obj *GitWebHookExecution) (runtime.Object, error)

type GitWebHookExecutionChangeHandlerFunc func(obj *GitWebHookExecution) (runtime.Object, error)

type GitWebHookExecutionLister interface {
	List(namespace string, selector labels.Selector) (ret []*GitWebHookExecution, err error)
	Get(namespace, name string) (*GitWebHookExecution, error)
}

type GitWebHookExecutionController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() GitWebHookExecutionLister
	AddHandler(ctx context.Context, name string, handler GitWebHookExecutionHandlerFunc)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, handler GitWebHookExecutionHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type GitWebHookExecutionInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*GitWebHookExecution) (*GitWebHookExecution, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*GitWebHookExecution, error)
	Get(name string, opts metav1.GetOptions) (*GitWebHookExecution, error)
	Update(*GitWebHookExecution) (*GitWebHookExecution, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*GitWebHookExecutionList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() GitWebHookExecutionController
	AddHandler(ctx context.Context, name string, sync GitWebHookExecutionHandlerFunc)
	AddLifecycle(ctx context.Context, name string, lifecycle GitWebHookExecutionLifecycle)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync GitWebHookExecutionHandlerFunc)
	AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle GitWebHookExecutionLifecycle)
}

type gitWebHookExecutionLister struct {
	controller *gitWebHookExecutionController
}

func (l *gitWebHookExecutionLister) List(namespace string, selector labels.Selector) (ret []*GitWebHookExecution, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*GitWebHookExecution))
	})
	return
}

func (l *gitWebHookExecutionLister) Get(namespace, name string) (*GitWebHookExecution, error) {
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
			Group:    GitWebHookExecutionGroupVersionKind.Group,
			Resource: "gitWebHookExecution",
		}, key)
	}
	return obj.(*GitWebHookExecution), nil
}

type gitWebHookExecutionController struct {
	controller.GenericController
}

func (c *gitWebHookExecutionController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *gitWebHookExecutionController) Lister() GitWebHookExecutionLister {
	return &gitWebHookExecutionLister{
		controller: c,
	}
}

func (c *gitWebHookExecutionController) AddHandler(ctx context.Context, name string, handler GitWebHookExecutionHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*GitWebHookExecution); ok {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

func (c *gitWebHookExecutionController) AddClusterScopedHandler(ctx context.Context, name, cluster string, handler GitWebHookExecutionHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*GitWebHookExecution); ok && controller.ObjectInCluster(cluster, obj) {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

type gitWebHookExecutionFactory struct {
}

func (c gitWebHookExecutionFactory) Object() runtime.Object {
	return &GitWebHookExecution{}
}

func (c gitWebHookExecutionFactory) List() runtime.Object {
	return &GitWebHookExecutionList{}
}

func (s *gitWebHookExecutionClient) Controller() GitWebHookExecutionController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.gitWebHookExecutionControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(GitWebHookExecutionGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &gitWebHookExecutionController{
		GenericController: genericController,
	}

	s.client.gitWebHookExecutionControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type gitWebHookExecutionClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   GitWebHookExecutionController
}

func (s *gitWebHookExecutionClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *gitWebHookExecutionClient) Create(o *GitWebHookExecution) (*GitWebHookExecution, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*GitWebHookExecution), err
}

func (s *gitWebHookExecutionClient) Get(name string, opts metav1.GetOptions) (*GitWebHookExecution, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*GitWebHookExecution), err
}

func (s *gitWebHookExecutionClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*GitWebHookExecution, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*GitWebHookExecution), err
}

func (s *gitWebHookExecutionClient) Update(o *GitWebHookExecution) (*GitWebHookExecution, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*GitWebHookExecution), err
}

func (s *gitWebHookExecutionClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *gitWebHookExecutionClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *gitWebHookExecutionClient) List(opts metav1.ListOptions) (*GitWebHookExecutionList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*GitWebHookExecutionList), err
}

func (s *gitWebHookExecutionClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *gitWebHookExecutionClient) Patch(o *GitWebHookExecution, patchType types.PatchType, data []byte, subresources ...string) (*GitWebHookExecution, error) {
	obj, err := s.objectClient.Patch(o.Name, o, patchType, data, subresources...)
	return obj.(*GitWebHookExecution), err
}

func (s *gitWebHookExecutionClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *gitWebHookExecutionClient) AddHandler(ctx context.Context, name string, sync GitWebHookExecutionHandlerFunc) {
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *gitWebHookExecutionClient) AddLifecycle(ctx context.Context, name string, lifecycle GitWebHookExecutionLifecycle) {
	sync := NewGitWebHookExecutionLifecycleAdapter(name, false, s, lifecycle)
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *gitWebHookExecutionClient) AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync GitWebHookExecutionHandlerFunc) {
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

func (s *gitWebHookExecutionClient) AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle GitWebHookExecutionLifecycle) {
	sync := NewGitWebHookExecutionLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

type GitWebHookExecutionIndexer func(obj *GitWebHookExecution) ([]string, error)

type GitWebHookExecutionClientCache interface {
	Get(namespace, name string) (*GitWebHookExecution, error)
	List(namespace string, selector labels.Selector) ([]*GitWebHookExecution, error)

	Index(name string, indexer GitWebHookExecutionIndexer)
	GetIndexed(name, key string) ([]*GitWebHookExecution, error)
}

type GitWebHookExecutionClient interface {
	Create(*GitWebHookExecution) (*GitWebHookExecution, error)
	Get(namespace, name string, opts metav1.GetOptions) (*GitWebHookExecution, error)
	Update(*GitWebHookExecution) (*GitWebHookExecution, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	List(namespace string, opts metav1.ListOptions) (*GitWebHookExecutionList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)

	Cache() GitWebHookExecutionClientCache

	OnCreate(ctx context.Context, name string, sync GitWebHookExecutionChangeHandlerFunc)
	OnChange(ctx context.Context, name string, sync GitWebHookExecutionChangeHandlerFunc)
	OnRemove(ctx context.Context, name string, sync GitWebHookExecutionChangeHandlerFunc)
	Enqueue(namespace, name string)

	Generic() controller.GenericController
	ObjectClient() *objectclient.ObjectClient
	Interface() GitWebHookExecutionInterface
}

type gitWebHookExecutionClientCache struct {
	client *gitWebHookExecutionClient2
}

type gitWebHookExecutionClient2 struct {
	iface      GitWebHookExecutionInterface
	controller GitWebHookExecutionController
}

func (n *gitWebHookExecutionClient2) Interface() GitWebHookExecutionInterface {
	return n.iface
}

func (n *gitWebHookExecutionClient2) Generic() controller.GenericController {
	return n.iface.Controller().Generic()
}

func (n *gitWebHookExecutionClient2) ObjectClient() *objectclient.ObjectClient {
	return n.Interface().ObjectClient()
}

func (n *gitWebHookExecutionClient2) Enqueue(namespace, name string) {
	n.iface.Controller().Enqueue(namespace, name)
}

func (n *gitWebHookExecutionClient2) Create(obj *GitWebHookExecution) (*GitWebHookExecution, error) {
	return n.iface.Create(obj)
}

func (n *gitWebHookExecutionClient2) Get(namespace, name string, opts metav1.GetOptions) (*GitWebHookExecution, error) {
	return n.iface.GetNamespaced(namespace, name, opts)
}

func (n *gitWebHookExecutionClient2) Update(obj *GitWebHookExecution) (*GitWebHookExecution, error) {
	return n.iface.Update(obj)
}

func (n *gitWebHookExecutionClient2) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	return n.iface.DeleteNamespaced(namespace, name, options)
}

func (n *gitWebHookExecutionClient2) List(namespace string, opts metav1.ListOptions) (*GitWebHookExecutionList, error) {
	return n.iface.List(opts)
}

func (n *gitWebHookExecutionClient2) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return n.iface.Watch(opts)
}

func (n *gitWebHookExecutionClientCache) Get(namespace, name string) (*GitWebHookExecution, error) {
	return n.client.controller.Lister().Get(namespace, name)
}

func (n *gitWebHookExecutionClientCache) List(namespace string, selector labels.Selector) ([]*GitWebHookExecution, error) {
	return n.client.controller.Lister().List(namespace, selector)
}

func (n *gitWebHookExecutionClient2) Cache() GitWebHookExecutionClientCache {
	n.loadController()
	return &gitWebHookExecutionClientCache{
		client: n,
	}
}

func (n *gitWebHookExecutionClient2) OnCreate(ctx context.Context, name string, sync GitWebHookExecutionChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-create", &gitWebHookExecutionLifecycleDelegate{create: sync})
}

func (n *gitWebHookExecutionClient2) OnChange(ctx context.Context, name string, sync GitWebHookExecutionChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-change", &gitWebHookExecutionLifecycleDelegate{update: sync})
}

func (n *gitWebHookExecutionClient2) OnRemove(ctx context.Context, name string, sync GitWebHookExecutionChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name, &gitWebHookExecutionLifecycleDelegate{remove: sync})
}

func (n *gitWebHookExecutionClientCache) Index(name string, indexer GitWebHookExecutionIndexer) {
	err := n.client.controller.Informer().GetIndexer().AddIndexers(map[string]cache.IndexFunc{
		name: func(obj interface{}) ([]string, error) {
			if v, ok := obj.(*GitWebHookExecution); ok {
				return indexer(v)
			}
			return nil, nil
		},
	})

	if err != nil {
		panic(err)
	}
}

func (n *gitWebHookExecutionClientCache) GetIndexed(name, key string) ([]*GitWebHookExecution, error) {
	var result []*GitWebHookExecution
	objs, err := n.client.controller.Informer().GetIndexer().ByIndex(name, key)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		if v, ok := obj.(*GitWebHookExecution); ok {
			result = append(result, v)
		}
	}

	return result, nil
}

func (n *gitWebHookExecutionClient2) loadController() {
	if n.controller == nil {
		n.controller = n.iface.Controller()
	}
}

type gitWebHookExecutionLifecycleDelegate struct {
	create GitWebHookExecutionChangeHandlerFunc
	update GitWebHookExecutionChangeHandlerFunc
	remove GitWebHookExecutionChangeHandlerFunc
}

func (n *gitWebHookExecutionLifecycleDelegate) HasCreate() bool {
	return n.create != nil
}

func (n *gitWebHookExecutionLifecycleDelegate) Create(obj *GitWebHookExecution) (runtime.Object, error) {
	if n.create == nil {
		return obj, nil
	}
	return n.create(obj)
}

func (n *gitWebHookExecutionLifecycleDelegate) HasFinalize() bool {
	return n.remove != nil
}

func (n *gitWebHookExecutionLifecycleDelegate) Remove(obj *GitWebHookExecution) (runtime.Object, error) {
	if n.remove == nil {
		return obj, nil
	}
	return n.remove(obj)
}

func (n *gitWebHookExecutionLifecycleDelegate) Updated(obj *GitWebHookExecution) (runtime.Object, error) {
	if n.update == nil {
		return obj, nil
	}
	return n.update(obj)
}
