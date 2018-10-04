package generator

var controllerTemplate = `package {{.schema.Version.Version}}

import (
	"context"

	{{.importPackage}}
	"github.com/rancher/norman/objectclient"
	"github.com/rancher/norman/controller"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

var (
	{{.schema.CodeName}}GroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "{{.schema.CodeName}}",
	}
	{{.schema.CodeName}}Resource = metav1.APIResource{
		Name:         "{{.schema.PluralName | toLower}}",
		SingularName: "{{.schema.ID | toLower}}",
{{- if eq .schema.Scope "namespace" }}
		Namespaced:   true,
{{ else }}
		Namespaced:   false,
{{- end }}
		Kind:         {{.schema.CodeName}}GroupVersionKind.Kind,
	}
)

type {{.schema.CodeName}}List struct {
	metav1.TypeMeta   %BACK%json:",inline"%BACK%
	metav1.ListMeta   %BACK%json:"metadata,omitempty"%BACK%
	Items             []{{.prefix}}{{.schema.CodeName}}
}

type {{.schema.CodeName}}HandlerFunc func(key string, obj *{{.prefix}}{{.schema.CodeName}}) error

type {{.schema.CodeName}}Lister interface {
	List(namespace string, selector labels.Selector) (ret []*{{.prefix}}{{.schema.CodeName}}, err error)
	Get(namespace, name string) (*{{.prefix}}{{.schema.CodeName}}, error)
}

type {{.schema.CodeName}}Controller interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() {{.schema.CodeName}}Lister
	AddHandler(name string, handler {{.schema.CodeName}}HandlerFunc)
	AddClusterScopedHandler(name, clusterName string, handler {{.schema.CodeName}}HandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type {{.schema.CodeName}}Interface interface {
    ObjectClient() *objectclient.ObjectClient
	Create(*{{.prefix}}{{.schema.CodeName}}) (*{{.prefix}}{{.schema.CodeName}}, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*{{.prefix}}{{.schema.CodeName}}, error)
	Get(name string, opts metav1.GetOptions) (*{{.prefix}}{{.schema.CodeName}}, error)
	Update(*{{.prefix}}{{.schema.CodeName}}) (*{{.prefix}}{{.schema.CodeName}}, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*{{.schema.CodeName}}List, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() {{.schema.CodeName}}Controller
	AddHandler(name string, sync {{.schema.CodeName}}HandlerFunc)
	AddLifecycle(name string, lifecycle {{.schema.CodeName}}Lifecycle)
	AddClusterScopedHandler(name, clusterName string, sync {{.schema.CodeName}}HandlerFunc)
	AddClusterScopedLifecycle(name, clusterName string, lifecycle {{.schema.CodeName}}Lifecycle)
}

type {{.schema.ID}}Lister struct {
	controller *{{.schema.ID}}Controller
}

func (l *{{.schema.ID}}Lister) List(namespace string, selector labels.Selector) (ret []*{{.prefix}}{{.schema.CodeName}}, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*{{.prefix}}{{.schema.CodeName}}))
	})
	return
}

func (l *{{.schema.ID}}Lister) Get(namespace, name string) (*{{.prefix}}{{.schema.CodeName}}, error) {
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
			Group: {{.schema.CodeName}}GroupVersionKind.Group,
			Resource: "{{.schema.ID}}",
		}, key)
	}
	return obj.(*{{.prefix}}{{.schema.CodeName}}), nil
}

type {{.schema.ID}}Controller struct {
	controller.GenericController
}

func (c *{{.schema.ID}}Controller) Generic() controller.GenericController {
	return c.GenericController
}

func (c *{{.schema.ID}}Controller) Lister() {{.schema.CodeName}}Lister {
	return &{{.schema.ID}}Lister{
		controller: c,
	}
}


func (c *{{.schema.ID}}Controller) AddHandler(name string, handler {{.schema.CodeName}}HandlerFunc) {
	c.GenericController.AddHandler(name, func(key string) error {
		obj, exists, err := c.Informer().GetStore().GetByKey(key)
		if err != nil {
			return err
		}
		if !exists {
			return handler(key, nil)
		}
		return handler(key, obj.(*{{.prefix}}{{.schema.CodeName}}))
	})
}

func (c *{{.schema.ID}}Controller) AddClusterScopedHandler(name, cluster string, handler {{.schema.CodeName}}HandlerFunc) {
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

		return handler(key, obj.(*{{.prefix}}{{.schema.CodeName}}))
	})
}

type {{.schema.ID}}Factory struct {
}

func (c {{.schema.ID}}Factory) Object() runtime.Object {
	return &{{.prefix}}{{.schema.CodeName}}{}
}

func (c {{.schema.ID}}Factory) List() runtime.Object {
	return &{{.schema.CodeName}}List{}
}

func (s *{{.schema.ID}}Client) Controller() {{.schema.CodeName}}Controller {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.{{.schema.ID}}Controllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController({{.schema.CodeName}}GroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &{{.schema.ID}}Controller{
		GenericController: genericController,
	}

	s.client.{{.schema.ID}}Controllers[s.ns] = c
    s.client.starters = append(s.client.starters, c)

	return c
}

type {{.schema.ID}}Client struct {
	client *Client
	ns string
	objectClient *objectclient.ObjectClient
	controller   {{.schema.CodeName}}Controller
}

func (s *{{.schema.ID}}Client) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *{{.schema.ID}}Client) Create(o *{{.prefix}}{{.schema.CodeName}}) (*{{.prefix}}{{.schema.CodeName}}, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*{{.prefix}}{{.schema.CodeName}}), err
}

func (s *{{.schema.ID}}Client) Get(name string, opts metav1.GetOptions) (*{{.prefix}}{{.schema.CodeName}}, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*{{.prefix}}{{.schema.CodeName}}), err
}

func (s *{{.schema.ID}}Client) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*{{.prefix}}{{.schema.CodeName}}, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*{{.prefix}}{{.schema.CodeName}}), err
}

func (s *{{.schema.ID}}Client) Update(o *{{.prefix}}{{.schema.CodeName}}) (*{{.prefix}}{{.schema.CodeName}}, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*{{.prefix}}{{.schema.CodeName}}), err
}

func (s *{{.schema.ID}}Client) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *{{.schema.ID}}Client) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *{{.schema.ID}}Client) List(opts metav1.ListOptions) (*{{.schema.CodeName}}List, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*{{.schema.CodeName}}List), err
}

func (s *{{.schema.ID}}Client) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *{{.schema.ID}}Client) Patch(o *{{.prefix}}{{.schema.CodeName}}, data []byte, subresources ...string) (*{{.prefix}}{{.schema.CodeName}}, error) {
	obj, err := s.objectClient.Patch(o.Name, o, data, subresources...)
	return obj.(*{{.prefix}}{{.schema.CodeName}}), err
}

func (s *{{.schema.ID}}Client) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *{{.schema.ID}}Client) AddHandler(name string, sync {{.schema.CodeName}}HandlerFunc) {
	s.Controller().AddHandler(name, sync)
}

func (s *{{.schema.ID}}Client) AddLifecycle(name string, lifecycle {{.schema.CodeName}}Lifecycle) {
	sync := New{{.schema.CodeName}}LifecycleAdapter(name, false, s, lifecycle)
	s.AddHandler(name, sync)
}

func (s *{{.schema.ID}}Client) AddClusterScopedHandler(name, clusterName string, sync {{.schema.CodeName}}HandlerFunc) {
	s.Controller().AddClusterScopedHandler(name, clusterName, sync)
}

func (s *{{.schema.ID}}Client) AddClusterScopedLifecycle(name, clusterName string, lifecycle {{.schema.CodeName}}Lifecycle) {
	sync := New{{.schema.CodeName}}LifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.AddClusterScopedHandler(name, clusterName, sync)
}
`
