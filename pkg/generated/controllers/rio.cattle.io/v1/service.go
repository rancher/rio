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

package v1

import (
	"context"

	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	clientset "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/rio.cattle.io/v1"
	informers "github.com/rancher/rio/pkg/generated/informers/externalversions/rio.cattle.io/v1"
	listers "github.com/rancher/rio/pkg/generated/listers/rio.cattle.io/v1"
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

type ServiceHandler func(string, *v1.Service) (*v1.Service, error)

type ServiceController interface {
	ServiceClient

	OnChange(ctx context.Context, name string, sync ServiceHandler)
	OnRemove(ctx context.Context, name string, sync ServiceHandler)
	Enqueue(namespace, name string)

	Cache() ServiceCache

	Informer() cache.SharedIndexInformer
	GroupVersionKind() schema.GroupVersionKind

	AddGenericHandler(ctx context.Context, name string, handler generic.Handler)
	AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler)
	Updater() generic.Updater
}

type ServiceClient interface {
	Create(*v1.Service) (*v1.Service, error)
	Update(*v1.Service) (*v1.Service, error)
	UpdateStatus(*v1.Service) (*v1.Service, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	Get(namespace, name string, options metav1.GetOptions) (*v1.Service, error)
	List(namespace string, opts metav1.ListOptions) (*v1.ServiceList, error)
	Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error)
	Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Service, err error)
}

type ServiceCache interface {
	Get(namespace, name string) (*v1.Service, error)
	List(namespace string, selector labels.Selector) ([]*v1.Service, error)

	AddIndexer(indexName string, indexer ServiceIndexer)
	GetByIndex(indexName, key string) ([]*v1.Service, error)
}

type ServiceIndexer func(obj *v1.Service) ([]string, error)

type serviceController struct {
	controllerManager *generic.ControllerManager
	clientGetter      clientset.ServicesGetter
	informer          informers.ServiceInformer
	gvk               schema.GroupVersionKind
}

func NewServiceController(gvk schema.GroupVersionKind, controllerManager *generic.ControllerManager, clientGetter clientset.ServicesGetter, informer informers.ServiceInformer) ServiceController {
	return &serviceController{
		controllerManager: controllerManager,
		clientGetter:      clientGetter,
		informer:          informer,
		gvk:               gvk,
	}
}

func FromServiceHandlerToHandler(sync ServiceHandler) generic.Handler {
	return func(key string, obj runtime.Object) (ret runtime.Object, err error) {
		var v *v1.Service
		if obj == nil {
			v, err = sync(key, nil)
		} else {
			v, err = sync(key, obj.(*v1.Service))
		}
		if v == nil {
			return nil, err
		}
		return v, err
	}
}

func (c *serviceController) Updater() generic.Updater {
	return func(obj runtime.Object) (runtime.Object, error) {
		newObj, err := c.Update(obj.(*v1.Service))
		if newObj == nil {
			return nil, err
		}
		return newObj, err
	}
}

func UpdateServiceOnChange(updater generic.Updater, handler ServiceHandler) ServiceHandler {
	return func(key string, obj *v1.Service) (*v1.Service, error) {
		if obj == nil {
			return handler(key, nil)
		}

		copyObj := obj.DeepCopy()
		newObj, err := handler(key, copyObj)
		if newObj != nil {
			copyObj = newObj
		}
		if obj.ResourceVersion == copyObj.ResourceVersion && !equality.Semantic.DeepEqual(obj, copyObj) {
			newObj, err := updater(copyObj)
			if newObj != nil && err == nil {
				copyObj = newObj.(*v1.Service)
			}
		}

		return copyObj, err
	}
}

func (c *serviceController) AddGenericHandler(ctx context.Context, name string, handler generic.Handler) {
	c.controllerManager.AddHandler(ctx, c.gvk, c.informer.Informer(), name, handler)
}

func (c *serviceController) AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler) {
	removeHandler := generic.NewRemoveHandler(name, c.Updater(), handler)
	c.controllerManager.AddHandler(ctx, c.gvk, c.informer.Informer(), name, removeHandler)
}

func (c *serviceController) OnChange(ctx context.Context, name string, sync ServiceHandler) {
	c.AddGenericHandler(ctx, name, FromServiceHandlerToHandler(sync))
}

func (c *serviceController) OnRemove(ctx context.Context, name string, sync ServiceHandler) {
	removeHandler := generic.NewRemoveHandler(name, c.Updater(), FromServiceHandlerToHandler(sync))
	c.AddGenericHandler(ctx, name, removeHandler)
}

func (c *serviceController) Enqueue(namespace, name string) {
	c.controllerManager.Enqueue(c.gvk, namespace, name)
}

func (c *serviceController) Informer() cache.SharedIndexInformer {
	return c.informer.Informer()
}

func (c *serviceController) GroupVersionKind() schema.GroupVersionKind {
	return c.gvk
}

func (c *serviceController) Cache() ServiceCache {
	return &serviceCache{
		lister:  c.informer.Lister(),
		indexer: c.informer.Informer().GetIndexer(),
	}
}

func (c *serviceController) Create(obj *v1.Service) (*v1.Service, error) {
	return c.clientGetter.Services(obj.Namespace).Create(obj)
}

func (c *serviceController) Update(obj *v1.Service) (*v1.Service, error) {
	return c.clientGetter.Services(obj.Namespace).Update(obj)
}

func (c *serviceController) UpdateStatus(obj *v1.Service) (*v1.Service, error) {
	return c.clientGetter.Services(obj.Namespace).UpdateStatus(obj)
}

func (c *serviceController) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	return c.clientGetter.Services(namespace).Delete(name, options)
}

func (c *serviceController) Get(namespace, name string, options metav1.GetOptions) (*v1.Service, error) {
	return c.clientGetter.Services(namespace).Get(name, options)
}

func (c *serviceController) List(namespace string, opts metav1.ListOptions) (*v1.ServiceList, error) {
	return c.clientGetter.Services(namespace).List(opts)
}

func (c *serviceController) Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	return c.clientGetter.Services(namespace).Watch(opts)
}

func (c *serviceController) Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Service, err error) {
	return c.clientGetter.Services(namespace).Patch(name, pt, data, subresources...)
}

type serviceCache struct {
	lister  listers.ServiceLister
	indexer cache.Indexer
}

func (c *serviceCache) Get(namespace, name string) (*v1.Service, error) {
	return c.lister.Services(namespace).Get(name)
}

func (c *serviceCache) List(namespace string, selector labels.Selector) ([]*v1.Service, error) {
	return c.lister.Services(namespace).List(selector)
}

func (c *serviceCache) AddIndexer(indexName string, indexer ServiceIndexer) {
	utilruntime.Must(c.indexer.AddIndexers(map[string]cache.IndexFunc{
		indexName: func(obj interface{}) (strings []string, e error) {
			return indexer(obj.(*v1.Service))
		},
	}))
}

func (c *serviceCache) GetByIndex(indexName, key string) (result []*v1.Service, err error) {
	objs, err := c.indexer.ByIndex(indexName, key)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		result = append(result, obj.(*v1.Service))
	}
	return result, nil
}
