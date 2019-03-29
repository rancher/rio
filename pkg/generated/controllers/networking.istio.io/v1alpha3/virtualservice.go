/*
Copyright 2019 Rancher Labs.

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

package v1alpha3

import (
	"context"

	"github.com/knative/pkg/apis/istio/v1alpha3"
	clientset "github.com/knative/pkg/client/clientset/versioned/typed/istio/v1alpha3"
	informers "github.com/knative/pkg/client/informers/externalversions/istio/v1alpha3"
	listers "github.com/knative/pkg/client/listers/istio/v1alpha3"
	"github.com/rancher/wrangler/pkg/generic"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

type VirtualServiceHandler func(string, *v1alpha3.VirtualService) (*v1alpha3.VirtualService, error)

type VirtualServiceController interface {
	VirtualServiceClient

	OnChange(ctx context.Context, name string, sync VirtualServiceHandler)
	OnRemove(ctx context.Context, name string, sync VirtualServiceHandler)
	Enqueue(namespace, name string)

	Cache() VirtualServiceCache

	Informer() cache.SharedIndexInformer
	GroupVersionKind() schema.GroupVersionKind

	AddGenericHandler(ctx context.Context, name string, handler generic.Handler)
	AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler)
	Updater() generic.Updater
}

type VirtualServiceClient interface {
	Create(*v1alpha3.VirtualService) (*v1alpha3.VirtualService, error)
	Update(*v1alpha3.VirtualService) (*v1alpha3.VirtualService, error)

	Delete(namespace, name string, options *metav1.DeleteOptions) error
	Get(namespace, name string, options metav1.GetOptions) (*v1alpha3.VirtualService, error)
	List(namespace string, opts metav1.ListOptions) (*v1alpha3.VirtualServiceList, error)
	Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error)
	Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha3.VirtualService, err error)
}

type VirtualServiceCache interface {
	Get(namespace, name string) (*v1alpha3.VirtualService, error)
	List(namespace string, selector labels.Selector) ([]*v1alpha3.VirtualService, error)

	AddIndexer(indexName string, indexer VirtualServiceIndexer)
	GetByIndex(indexName, key string) ([]*v1alpha3.VirtualService, error)
}

type VirtualServiceIndexer func(obj *v1alpha3.VirtualService) ([]string, error)

type virtualServiceController struct {
	controllerManager *generic.ControllerManager
	clientGetter      clientset.VirtualServicesGetter
	informer          informers.VirtualServiceInformer
	gvk               schema.GroupVersionKind
}

func NewVirtualServiceController(gvk schema.GroupVersionKind, controllerManager *generic.ControllerManager, clientGetter clientset.VirtualServicesGetter, informer informers.VirtualServiceInformer) VirtualServiceController {
	return &virtualServiceController{
		controllerManager: controllerManager,
		clientGetter:      clientGetter,
		informer:          informer,
		gvk:               gvk,
	}
}

func FromVirtualServiceHandlerToHandler(sync VirtualServiceHandler) generic.Handler {
	return func(key string, obj runtime.Object) (ret runtime.Object, err error) {
		var v *v1alpha3.VirtualService
		if obj == nil {
			v, err = sync(key, nil)
		} else {
			v, err = sync(key, obj.(*v1alpha3.VirtualService))
		}
		if v == nil {
			return nil, err
		}
		return v, err
	}
}

func (c *virtualServiceController) Updater() generic.Updater {
	return func(obj runtime.Object) (runtime.Object, error) {
		newObj, err := c.Update(obj.(*v1alpha3.VirtualService))
		if newObj == nil {
			return nil, err
		}
		return newObj, err
	}
}

func UpdateVirtualServiceOnChange(updater generic.Updater, handler VirtualServiceHandler) VirtualServiceHandler {
	return func(key string, obj *v1alpha3.VirtualService) (*v1alpha3.VirtualService, error) {
		if obj == nil {
			return handler(key, nil)
		}

		copyObj := obj.DeepCopy()
		newObj, err := handler(key, copyObj)
		if newObj != nil {
			copyObj = newObj
		}
		if obj.ResourceVersion == copyObj.ResourceVersion && !equality.Semantic.DeepEqual(obj, copyObj) {
			newObj, _ := updater(copyObj)
			if newObj != nil {
				copyObj = newObj.(*v1alpha3.VirtualService)
			}
		}

		return copyObj, err
	}
}

func (c *virtualServiceController) AddGenericHandler(ctx context.Context, name string, handler generic.Handler) {
	c.controllerManager.AddHandler(ctx, c.gvk, c.informer.Informer(), name, handler)
}

func (c *virtualServiceController) AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler) {
	removeHandler := generic.NewRemoveHandler(name, c.Updater(), handler)
	c.controllerManager.AddHandler(ctx, c.gvk, c.informer.Informer(), name, removeHandler)
}

func (c *virtualServiceController) OnChange(ctx context.Context, name string, sync VirtualServiceHandler) {
	c.AddGenericHandler(ctx, name, FromVirtualServiceHandlerToHandler(sync))
}

func (c *virtualServiceController) OnRemove(ctx context.Context, name string, sync VirtualServiceHandler) {
	removeHandler := generic.NewRemoveHandler(name, c.Updater(), FromVirtualServiceHandlerToHandler(sync))
	c.AddGenericHandler(ctx, name, removeHandler)
}

func (c *virtualServiceController) Enqueue(namespace, name string) {
	c.controllerManager.Enqueue(c.gvk, namespace, name)
}

func (c *virtualServiceController) Informer() cache.SharedIndexInformer {
	return c.informer.Informer()
}

func (c *virtualServiceController) GroupVersionKind() schema.GroupVersionKind {
	return c.gvk
}

func (c *virtualServiceController) Cache() VirtualServiceCache {
	return &virtualServiceCache{
		lister:  c.informer.Lister(),
		indexer: c.informer.Informer().GetIndexer(),
	}
}

func (c *virtualServiceController) Create(obj *v1alpha3.VirtualService) (*v1alpha3.VirtualService, error) {
	return c.clientGetter.VirtualServices(obj.Namespace).Create(obj)
}

func (c *virtualServiceController) Update(obj *v1alpha3.VirtualService) (*v1alpha3.VirtualService, error) {
	return c.clientGetter.VirtualServices(obj.Namespace).Update(obj)
}

func (c *virtualServiceController) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	return c.clientGetter.VirtualServices(namespace).Delete(name, options)
}

func (c *virtualServiceController) Get(namespace, name string, options metav1.GetOptions) (*v1alpha3.VirtualService, error) {
	return c.clientGetter.VirtualServices(namespace).Get(name, options)
}

func (c *virtualServiceController) List(namespace string, opts metav1.ListOptions) (*v1alpha3.VirtualServiceList, error) {
	return c.clientGetter.VirtualServices(namespace).List(opts)
}

func (c *virtualServiceController) Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	return c.clientGetter.VirtualServices(namespace).Watch(opts)
}

func (c *virtualServiceController) Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha3.VirtualService, err error) {
	return c.clientGetter.VirtualServices(namespace).Patch(name, pt, data, subresources...)
}

type virtualServiceCache struct {
	lister  listers.VirtualServiceLister
	indexer cache.Indexer
}

func (c *virtualServiceCache) Get(namespace, name string) (*v1alpha3.VirtualService, error) {
	return c.lister.VirtualServices(namespace).Get(name)
}

func (c *virtualServiceCache) List(namespace string, selector labels.Selector) ([]*v1alpha3.VirtualService, error) {
	return c.lister.VirtualServices(namespace).List(selector)
}

func (c *virtualServiceCache) AddIndexer(indexName string, indexer VirtualServiceIndexer) {
	utilruntime.Must(c.indexer.AddIndexers(map[string]cache.IndexFunc{
		indexName: func(obj interface{}) (strings []string, e error) {
			return indexer(obj.(*v1alpha3.VirtualService))
		},
	}))
}

func (c *virtualServiceCache) GetByIndex(indexName, key string) (result []*v1alpha3.VirtualService, err error) {
	objs, err := c.indexer.ByIndex(indexName, key)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		result = append(result, obj.(*v1alpha3.VirtualService))
	}
	return result, nil
}
