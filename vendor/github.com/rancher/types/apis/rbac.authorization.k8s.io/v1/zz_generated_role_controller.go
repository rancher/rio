package v1

import (
	"context"

	"github.com/rancher/norman/controller"
	"github.com/rancher/norman/objectclient"
	"k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

var (
	RoleGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "Role",
	}
	RoleResource = metav1.APIResource{
		Name:         "roles",
		SingularName: "role",
		Namespaced:   true,

		Kind: RoleGroupVersionKind.Kind,
	}
)

type RoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []v1.Role
}

type RoleHandlerFunc func(key string, obj *v1.Role) error

type RoleLister interface {
	List(namespace string, selector labels.Selector) (ret []*v1.Role, err error)
	Get(namespace, name string) (*v1.Role, error)
}

type RoleController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() RoleLister
	AddHandler(name string, handler RoleHandlerFunc)
	AddClusterScopedHandler(name, clusterName string, handler RoleHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type RoleInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*v1.Role) (*v1.Role, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*v1.Role, error)
	Get(name string, opts metav1.GetOptions) (*v1.Role, error)
	Update(*v1.Role) (*v1.Role, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*RoleList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() RoleController
	AddHandler(name string, sync RoleHandlerFunc)
	AddLifecycle(name string, lifecycle RoleLifecycle)
	AddClusterScopedHandler(name, clusterName string, sync RoleHandlerFunc)
	AddClusterScopedLifecycle(name, clusterName string, lifecycle RoleLifecycle)
}

type roleLister struct {
	controller *roleController
}

func (l *roleLister) List(namespace string, selector labels.Selector) (ret []*v1.Role, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*v1.Role))
	})
	return
}

func (l *roleLister) Get(namespace, name string) (*v1.Role, error) {
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
			Group:    RoleGroupVersionKind.Group,
			Resource: "role",
		}, key)
	}
	return obj.(*v1.Role), nil
}

type roleController struct {
	controller.GenericController
}

func (c *roleController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *roleController) Lister() RoleLister {
	return &roleLister{
		controller: c,
	}
}

func (c *roleController) AddHandler(name string, handler RoleHandlerFunc) {
	c.GenericController.AddHandler(name, func(key string) error {
		obj, exists, err := c.Informer().GetStore().GetByKey(key)
		if err != nil {
			return err
		}
		if !exists {
			return handler(key, nil)
		}
		return handler(key, obj.(*v1.Role))
	})
}

func (c *roleController) AddClusterScopedHandler(name, cluster string, handler RoleHandlerFunc) {
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

		return handler(key, obj.(*v1.Role))
	})
}

type roleFactory struct {
}

func (c roleFactory) Object() runtime.Object {
	return &v1.Role{}
}

func (c roleFactory) List() runtime.Object {
	return &RoleList{}
}

func (s *roleClient) Controller() RoleController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.roleControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(RoleGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &roleController{
		GenericController: genericController,
	}

	s.client.roleControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type roleClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   RoleController
}

func (s *roleClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *roleClient) Create(o *v1.Role) (*v1.Role, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*v1.Role), err
}

func (s *roleClient) Get(name string, opts metav1.GetOptions) (*v1.Role, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*v1.Role), err
}

func (s *roleClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*v1.Role, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*v1.Role), err
}

func (s *roleClient) Update(o *v1.Role) (*v1.Role, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*v1.Role), err
}

func (s *roleClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *roleClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *roleClient) List(opts metav1.ListOptions) (*RoleList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*RoleList), err
}

func (s *roleClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *roleClient) Patch(o *v1.Role, data []byte, subresources ...string) (*v1.Role, error) {
	obj, err := s.objectClient.Patch(o.Name, o, data, subresources...)
	return obj.(*v1.Role), err
}

func (s *roleClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *roleClient) AddHandler(name string, sync RoleHandlerFunc) {
	s.Controller().AddHandler(name, sync)
}

func (s *roleClient) AddLifecycle(name string, lifecycle RoleLifecycle) {
	sync := NewRoleLifecycleAdapter(name, false, s, lifecycle)
	s.AddHandler(name, sync)
}

func (s *roleClient) AddClusterScopedHandler(name, clusterName string, sync RoleHandlerFunc) {
	s.Controller().AddClusterScopedHandler(name, clusterName, sync)
}

func (s *roleClient) AddClusterScopedLifecycle(name, clusterName string, lifecycle RoleLifecycle) {
	sync := NewRoleLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.AddClusterScopedHandler(name, clusterName, sync)
}
