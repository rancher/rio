/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by main. DO NOT EDIT.

package v1

import (
	"context"
	"time"

	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/condition"
	"github.com/rancher/wrangler/pkg/generic"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	informers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes/typed/core/v1"
	listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type NodeHandler func(string, *v1.Node) (*v1.Node, error)

type NodeController interface {
	generic.ControllerMeta
	NodeClient

	OnChange(ctx context.Context, name string, sync NodeHandler)
	OnRemove(ctx context.Context, name string, sync NodeHandler)
	Enqueue(name string)
	EnqueueAfter(name string, duration time.Duration)

	Cache() NodeCache
}

type NodeClient interface {
	Create(*v1.Node) (*v1.Node, error)
	Update(*v1.Node) (*v1.Node, error)
	UpdateStatus(*v1.Node) (*v1.Node, error)
	Delete(name string, options *metav1.DeleteOptions) error
	Get(name string, options metav1.GetOptions) (*v1.Node, error)
	List(opts metav1.ListOptions) (*v1.NodeList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Node, err error)
}

type NodeCache interface {
	Get(name string) (*v1.Node, error)
	List(selector labels.Selector) ([]*v1.Node, error)

	AddIndexer(indexName string, indexer NodeIndexer)
	GetByIndex(indexName, key string) ([]*v1.Node, error)
}

type NodeIndexer func(obj *v1.Node) ([]string, error)

type nodeController struct {
	controllerManager *generic.ControllerManager
	clientGetter      clientset.NodesGetter
	informer          informers.NodeInformer
	gvk               schema.GroupVersionKind
}

func NewNodeController(gvk schema.GroupVersionKind, controllerManager *generic.ControllerManager, clientGetter clientset.NodesGetter, informer informers.NodeInformer) NodeController {
	return &nodeController{
		controllerManager: controllerManager,
		clientGetter:      clientGetter,
		informer:          informer,
		gvk:               gvk,
	}
}

func FromNodeHandlerToHandler(sync NodeHandler) generic.Handler {
	return func(key string, obj runtime.Object) (ret runtime.Object, err error) {
		var v *v1.Node
		if obj == nil {
			v, err = sync(key, nil)
		} else {
			v, err = sync(key, obj.(*v1.Node))
		}
		if v == nil {
			return nil, err
		}
		return v, err
	}
}

func (c *nodeController) Updater() generic.Updater {
	return func(obj runtime.Object) (runtime.Object, error) {
		newObj, err := c.Update(obj.(*v1.Node))
		if newObj == nil {
			return nil, err
		}
		return newObj, err
	}
}

func UpdateNodeDeepCopyOnChange(client NodeClient, obj *v1.Node, handler func(obj *v1.Node) (*v1.Node, error)) (*v1.Node, error) {
	if obj == nil {
		return obj, nil
	}

	copyObj := obj.DeepCopy()
	newObj, err := handler(copyObj)
	if newObj != nil {
		copyObj = newObj
	}
	if obj.ResourceVersion == copyObj.ResourceVersion && !equality.Semantic.DeepEqual(obj, copyObj) {
		return client.Update(copyObj)
	}

	return copyObj, err
}

func (c *nodeController) AddGenericHandler(ctx context.Context, name string, handler generic.Handler) {
	c.controllerManager.AddHandler(ctx, c.gvk, c.informer.Informer(), name, handler)
}

func (c *nodeController) AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler) {
	removeHandler := generic.NewRemoveHandler(name, c.Updater(), handler)
	c.controllerManager.AddHandler(ctx, c.gvk, c.informer.Informer(), name, removeHandler)
}

func (c *nodeController) OnChange(ctx context.Context, name string, sync NodeHandler) {
	c.AddGenericHandler(ctx, name, FromNodeHandlerToHandler(sync))
}

func (c *nodeController) OnRemove(ctx context.Context, name string, sync NodeHandler) {
	removeHandler := generic.NewRemoveHandler(name, c.Updater(), FromNodeHandlerToHandler(sync))
	c.AddGenericHandler(ctx, name, removeHandler)
}

func (c *nodeController) Enqueue(name string) {
	c.controllerManager.Enqueue(c.gvk, c.informer.Informer(), "", name)
}

func (c *nodeController) EnqueueAfter(name string, duration time.Duration) {
	c.controllerManager.EnqueueAfter(c.gvk, c.informer.Informer(), "", name, duration)
}

func (c *nodeController) Informer() cache.SharedIndexInformer {
	return c.informer.Informer()
}

func (c *nodeController) GroupVersionKind() schema.GroupVersionKind {
	return c.gvk
}

func (c *nodeController) Cache() NodeCache {
	return &nodeCache{
		lister:  c.informer.Lister(),
		indexer: c.informer.Informer().GetIndexer(),
	}
}

func (c *nodeController) Create(obj *v1.Node) (*v1.Node, error) {
	return c.clientGetter.Nodes().Create(obj)
}

func (c *nodeController) Update(obj *v1.Node) (*v1.Node, error) {
	return c.clientGetter.Nodes().Update(obj)
}

func (c *nodeController) UpdateStatus(obj *v1.Node) (*v1.Node, error) {
	return c.clientGetter.Nodes().UpdateStatus(obj)
}

func (c *nodeController) Delete(name string, options *metav1.DeleteOptions) error {
	return c.clientGetter.Nodes().Delete(name, options)
}

func (c *nodeController) Get(name string, options metav1.GetOptions) (*v1.Node, error) {
	return c.clientGetter.Nodes().Get(name, options)
}

func (c *nodeController) List(opts metav1.ListOptions) (*v1.NodeList, error) {
	return c.clientGetter.Nodes().List(opts)
}

func (c *nodeController) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return c.clientGetter.Nodes().Watch(opts)
}

func (c *nodeController) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Node, err error) {
	return c.clientGetter.Nodes().Patch(name, pt, data, subresources...)
}

type nodeCache struct {
	lister  listers.NodeLister
	indexer cache.Indexer
}

func (c *nodeCache) Get(name string) (*v1.Node, error) {
	return c.lister.Get(name)
}

func (c *nodeCache) List(selector labels.Selector) ([]*v1.Node, error) {
	return c.lister.List(selector)
}

func (c *nodeCache) AddIndexer(indexName string, indexer NodeIndexer) {
	utilruntime.Must(c.indexer.AddIndexers(map[string]cache.IndexFunc{
		indexName: func(obj interface{}) (strings []string, e error) {
			return indexer(obj.(*v1.Node))
		},
	}))
}

func (c *nodeCache) GetByIndex(indexName, key string) (result []*v1.Node, err error) {
	objs, err := c.indexer.ByIndex(indexName, key)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		result = append(result, obj.(*v1.Node))
	}
	return result, nil
}

type NodeStatusHandler func(obj *v1.Node, status v1.NodeStatus) (v1.NodeStatus, error)

type NodeGeneratingHandler func(obj *v1.Node, status v1.NodeStatus) ([]runtime.Object, v1.NodeStatus, error)

func RegisterNodeStatusHandler(ctx context.Context, controller NodeController, condition condition.Cond, name string, handler NodeStatusHandler) {
	statusHandler := &nodeStatusHandler{
		client:    controller,
		condition: condition,
		handler:   handler,
	}
	controller.AddGenericHandler(ctx, name, FromNodeHandlerToHandler(statusHandler.sync))
}

func RegisterNodeGeneratingHandler(ctx context.Context, controller NodeController, apply apply.Apply,
	condition condition.Cond, name string, handler NodeGeneratingHandler, opts *generic.GeneratingHandlerOptions) {
	statusHandler := &nodeGeneratingHandler{
		NodeGeneratingHandler: handler,
		apply:                 apply,
		name:                  name,
		gvk:                   controller.GroupVersionKind(),
	}
	if opts != nil {
		statusHandler.opts = *opts
	}
	RegisterNodeStatusHandler(ctx, controller, condition, name, statusHandler.Handle)
}

type nodeStatusHandler struct {
	client    NodeClient
	condition condition.Cond
	handler   NodeStatusHandler
}

func (a *nodeStatusHandler) sync(key string, obj *v1.Node) (*v1.Node, error) {
	if obj == nil {
		return obj, nil
	}

	status := obj.Status
	obj = obj.DeepCopy()
	newStatus, err := a.handler(obj, obj.Status)
	if err != nil {
		// Revert to old status on error
		newStatus = *status.DeepCopy()
	}

	if a.condition != "" {
		if errors.IsConflict(err) {
			a.condition.SetError(obj, "", nil)
		} else {
			a.condition.SetError(obj, "", err)
		}
	}
	if !equality.Semantic.DeepEqual(status, newStatus) {
		var newErr error
		obj.Status = newStatus
		obj, newErr = a.client.UpdateStatus(obj)
		if err == nil {
			err = newErr
		}
	}
	return obj, err
}

type nodeGeneratingHandler struct {
	NodeGeneratingHandler
	apply apply.Apply
	opts  generic.GeneratingHandlerOptions
	gvk   schema.GroupVersionKind
	name  string
}

func (a *nodeGeneratingHandler) Handle(obj *v1.Node, status v1.NodeStatus) (v1.NodeStatus, error) {
	objs, newStatus, err := a.NodeGeneratingHandler(obj, status)
	if err != nil {
		return newStatus, err
	}

	apply := a.apply

	if !a.opts.DynamicLookup {
		apply = apply.WithStrictCaching()
	}

	if !a.opts.AllowCrossNamespace && !a.opts.AllowClusterScoped {
		apply = apply.WithSetOwnerReference(true, false).
			WithDefaultNamespace(obj.GetNamespace()).
			WithListerNamespace(obj.GetNamespace())
	}

	if !a.opts.AllowClusterScoped {
		apply = apply.WithRestrictClusterScoped()
	}

	return newStatus, apply.
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects(objs...)
}
