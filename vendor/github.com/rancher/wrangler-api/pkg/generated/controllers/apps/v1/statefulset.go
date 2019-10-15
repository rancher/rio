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
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	informers "k8s.io/client-go/informers/apps/v1"
	clientset "k8s.io/client-go/kubernetes/typed/apps/v1"
	listers "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
)

type StatefulSetHandler func(string, *v1.StatefulSet) (*v1.StatefulSet, error)

type StatefulSetController interface {
	generic.ControllerMeta
	StatefulSetClient

	OnChange(ctx context.Context, name string, sync StatefulSetHandler)
	OnRemove(ctx context.Context, name string, sync StatefulSetHandler)
	Enqueue(namespace, name string)
	EnqueueAfter(namespace, name string, duration time.Duration)

	Cache() StatefulSetCache
}

type StatefulSetClient interface {
	Create(*v1.StatefulSet) (*v1.StatefulSet, error)
	Update(*v1.StatefulSet) (*v1.StatefulSet, error)
	UpdateStatus(*v1.StatefulSet) (*v1.StatefulSet, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	Get(namespace, name string, options metav1.GetOptions) (*v1.StatefulSet, error)
	List(namespace string, opts metav1.ListOptions) (*v1.StatefulSetList, error)
	Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error)
	Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.StatefulSet, err error)
}

type StatefulSetCache interface {
	Get(namespace, name string) (*v1.StatefulSet, error)
	List(namespace string, selector labels.Selector) ([]*v1.StatefulSet, error)

	AddIndexer(indexName string, indexer StatefulSetIndexer)
	GetByIndex(indexName, key string) ([]*v1.StatefulSet, error)
}

type StatefulSetIndexer func(obj *v1.StatefulSet) ([]string, error)

type statefulSetController struct {
	controllerManager *generic.ControllerManager
	clientGetter      clientset.StatefulSetsGetter
	informer          informers.StatefulSetInformer
	gvk               schema.GroupVersionKind
}

func NewStatefulSetController(gvk schema.GroupVersionKind, controllerManager *generic.ControllerManager, clientGetter clientset.StatefulSetsGetter, informer informers.StatefulSetInformer) StatefulSetController {
	return &statefulSetController{
		controllerManager: controllerManager,
		clientGetter:      clientGetter,
		informer:          informer,
		gvk:               gvk,
	}
}

func FromStatefulSetHandlerToHandler(sync StatefulSetHandler) generic.Handler {
	return func(key string, obj runtime.Object) (ret runtime.Object, err error) {
		var v *v1.StatefulSet
		if obj == nil {
			v, err = sync(key, nil)
		} else {
			v, err = sync(key, obj.(*v1.StatefulSet))
		}
		if v == nil {
			return nil, err
		}
		return v, err
	}
}

func (c *statefulSetController) Updater() generic.Updater {
	return func(obj runtime.Object) (runtime.Object, error) {
		newObj, err := c.Update(obj.(*v1.StatefulSet))
		if newObj == nil {
			return nil, err
		}
		return newObj, err
	}
}

func UpdateStatefulSetDeepCopyOnChange(client StatefulSetClient, obj *v1.StatefulSet, handler func(obj *v1.StatefulSet) (*v1.StatefulSet, error)) (*v1.StatefulSet, error) {
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

func (c *statefulSetController) AddGenericHandler(ctx context.Context, name string, handler generic.Handler) {
	c.controllerManager.AddHandler(ctx, c.gvk, c.informer.Informer(), name, handler)
}

func (c *statefulSetController) AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler) {
	removeHandler := generic.NewRemoveHandler(name, c.Updater(), handler)
	c.controllerManager.AddHandler(ctx, c.gvk, c.informer.Informer(), name, removeHandler)
}

func (c *statefulSetController) OnChange(ctx context.Context, name string, sync StatefulSetHandler) {
	c.AddGenericHandler(ctx, name, FromStatefulSetHandlerToHandler(sync))
}

func (c *statefulSetController) OnRemove(ctx context.Context, name string, sync StatefulSetHandler) {
	removeHandler := generic.NewRemoveHandler(name, c.Updater(), FromStatefulSetHandlerToHandler(sync))
	c.AddGenericHandler(ctx, name, removeHandler)
}

func (c *statefulSetController) Enqueue(namespace, name string) {
	c.controllerManager.Enqueue(c.gvk, c.informer.Informer(), namespace, name)
}

func (c *statefulSetController) EnqueueAfter(namespace, name string, duration time.Duration) {
	c.controllerManager.EnqueueAfter(c.gvk, c.informer.Informer(), namespace, name, duration)
}

func (c *statefulSetController) Informer() cache.SharedIndexInformer {
	return c.informer.Informer()
}

func (c *statefulSetController) GroupVersionKind() schema.GroupVersionKind {
	return c.gvk
}

func (c *statefulSetController) Cache() StatefulSetCache {
	return &statefulSetCache{
		lister:  c.informer.Lister(),
		indexer: c.informer.Informer().GetIndexer(),
	}
}

func (c *statefulSetController) Create(obj *v1.StatefulSet) (*v1.StatefulSet, error) {
	return c.clientGetter.StatefulSets(obj.Namespace).Create(obj)
}

func (c *statefulSetController) Update(obj *v1.StatefulSet) (*v1.StatefulSet, error) {
	return c.clientGetter.StatefulSets(obj.Namespace).Update(obj)
}

func (c *statefulSetController) UpdateStatus(obj *v1.StatefulSet) (*v1.StatefulSet, error) {
	return c.clientGetter.StatefulSets(obj.Namespace).UpdateStatus(obj)
}

func (c *statefulSetController) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	return c.clientGetter.StatefulSets(namespace).Delete(name, options)
}

func (c *statefulSetController) Get(namespace, name string, options metav1.GetOptions) (*v1.StatefulSet, error) {
	return c.clientGetter.StatefulSets(namespace).Get(name, options)
}

func (c *statefulSetController) List(namespace string, opts metav1.ListOptions) (*v1.StatefulSetList, error) {
	return c.clientGetter.StatefulSets(namespace).List(opts)
}

func (c *statefulSetController) Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	return c.clientGetter.StatefulSets(namespace).Watch(opts)
}

func (c *statefulSetController) Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.StatefulSet, err error) {
	return c.clientGetter.StatefulSets(namespace).Patch(name, pt, data, subresources...)
}

type statefulSetCache struct {
	lister  listers.StatefulSetLister
	indexer cache.Indexer
}

func (c *statefulSetCache) Get(namespace, name string) (*v1.StatefulSet, error) {
	return c.lister.StatefulSets(namespace).Get(name)
}

func (c *statefulSetCache) List(namespace string, selector labels.Selector) ([]*v1.StatefulSet, error) {
	return c.lister.StatefulSets(namespace).List(selector)
}

func (c *statefulSetCache) AddIndexer(indexName string, indexer StatefulSetIndexer) {
	utilruntime.Must(c.indexer.AddIndexers(map[string]cache.IndexFunc{
		indexName: func(obj interface{}) (strings []string, e error) {
			return indexer(obj.(*v1.StatefulSet))
		},
	}))
}

func (c *statefulSetCache) GetByIndex(indexName, key string) (result []*v1.StatefulSet, err error) {
	objs, err := c.indexer.ByIndex(indexName, key)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		result = append(result, obj.(*v1.StatefulSet))
	}
	return result, nil
}

type StatefulSetStatusHandler func(obj *v1.StatefulSet, status v1.StatefulSetStatus) (v1.StatefulSetStatus, error)

type StatefulSetGeneratingHandler func(obj *v1.StatefulSet, status v1.StatefulSetStatus) ([]runtime.Object, v1.StatefulSetStatus, error)

func RegisterStatefulSetStatusHandler(ctx context.Context, controller StatefulSetController, condition condition.Cond, name string, handler StatefulSetStatusHandler) {
	statusHandler := &statefulSetStatusHandler{
		client:    controller,
		condition: condition,
		handler:   handler,
	}
	controller.AddGenericHandler(ctx, name, FromStatefulSetHandlerToHandler(statusHandler.sync))
}

func RegisterStatefulSetGeneratingHandler(ctx context.Context, controller StatefulSetController, apply apply.Apply,
	condition condition.Cond, name string, handler StatefulSetGeneratingHandler, opts *generic.GeneratingHandlerOptions) {
	statusHandler := &statefulSetGeneratingHandler{
		StatefulSetGeneratingHandler: handler,
		apply:                        apply,
		name:                         name,
		gvk:                          controller.GroupVersionKind(),
	}
	if opts != nil {
		statusHandler.opts = *opts
	}
	RegisterStatefulSetStatusHandler(ctx, controller, condition, name, statusHandler.Handle)
}

type statefulSetStatusHandler struct {
	client    StatefulSetClient
	condition condition.Cond
	handler   StatefulSetStatusHandler
}

func (a *statefulSetStatusHandler) sync(key string, obj *v1.StatefulSet) (*v1.StatefulSet, error) {
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

type statefulSetGeneratingHandler struct {
	StatefulSetGeneratingHandler
	apply apply.Apply
	opts  generic.GeneratingHandlerOptions
	gvk   schema.GroupVersionKind
	name  string
}

func (a *statefulSetGeneratingHandler) Handle(obj *v1.StatefulSet, status v1.StatefulSetStatus) (v1.StatefulSetStatus, error) {
	objs, newStatus, err := a.StatefulSetGeneratingHandler(obj, status)
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
