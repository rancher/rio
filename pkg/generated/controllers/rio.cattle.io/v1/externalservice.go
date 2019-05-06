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

type ExternalServiceHandler func(string, *v1.ExternalService) (*v1.ExternalService, error)

type ExternalServiceController interface {
	ExternalServiceClient

	OnChange(ctx context.Context, name string, sync ExternalServiceHandler)
	OnRemove(ctx context.Context, name string, sync ExternalServiceHandler)
	Enqueue(namespace, name string)

	Cache() ExternalServiceCache

	Informer() cache.SharedIndexInformer
	GroupVersionKind() schema.GroupVersionKind

	AddGenericHandler(ctx context.Context, name string, handler generic.Handler)
	AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler)
	Updater() generic.Updater
}

type ExternalServiceClient interface {
	Create(*v1.ExternalService) (*v1.ExternalService, error)
	Update(*v1.ExternalService) (*v1.ExternalService, error)
	UpdateStatus(*v1.ExternalService) (*v1.ExternalService, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	Get(namespace, name string, options metav1.GetOptions) (*v1.ExternalService, error)
	List(namespace string, opts metav1.ListOptions) (*v1.ExternalServiceList, error)
	Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error)
	Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.ExternalService, err error)
}

type ExternalServiceCache interface {
	Get(namespace, name string) (*v1.ExternalService, error)
	List(namespace string, selector labels.Selector) ([]*v1.ExternalService, error)

	AddIndexer(indexName string, indexer ExternalServiceIndexer)
	GetByIndex(indexName, key string) ([]*v1.ExternalService, error)
}

type ExternalServiceIndexer func(obj *v1.ExternalService) ([]string, error)

type externalServiceController struct {
	controllerManager *generic.ControllerManager
	clientGetter      clientset.ExternalServicesGetter
	informer          informers.ExternalServiceInformer
	gvk               schema.GroupVersionKind
}

func NewExternalServiceController(gvk schema.GroupVersionKind, controllerManager *generic.ControllerManager, clientGetter clientset.ExternalServicesGetter, informer informers.ExternalServiceInformer) ExternalServiceController {
	return &externalServiceController{
		controllerManager: controllerManager,
		clientGetter:      clientGetter,
		informer:          informer,
		gvk:               gvk,
	}
}

func FromExternalServiceHandlerToHandler(sync ExternalServiceHandler) generic.Handler {
	return func(key string, obj runtime.Object) (ret runtime.Object, err error) {
		var v *v1.ExternalService
		if obj == nil {
			v, err = sync(key, nil)
		} else {
			v, err = sync(key, obj.(*v1.ExternalService))
		}
		if v == nil {
			return nil, err
		}
		return v, err
	}
}

func (c *externalServiceController) Updater() generic.Updater {
	return func(obj runtime.Object) (runtime.Object, error) {
		newObj, err := c.Update(obj.(*v1.ExternalService))
		if newObj == nil {
			return nil, err
		}
		return newObj, err
	}
}

func UpdateExternalServiceOnChange(updater generic.Updater, handler ExternalServiceHandler) ExternalServiceHandler {
	return func(key string, obj *v1.ExternalService) (*v1.ExternalService, error) {
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
				copyObj = newObj.(*v1.ExternalService)
			}
		}

		return copyObj, err
	}
}

func (c *externalServiceController) AddGenericHandler(ctx context.Context, name string, handler generic.Handler) {
	c.controllerManager.AddHandler(ctx, c.gvk, c.informer.Informer(), name, handler)
}

func (c *externalServiceController) AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler) {
	removeHandler := generic.NewRemoveHandler(name, c.Updater(), handler)
	c.controllerManager.AddHandler(ctx, c.gvk, c.informer.Informer(), name, removeHandler)
}

func (c *externalServiceController) OnChange(ctx context.Context, name string, sync ExternalServiceHandler) {
	c.AddGenericHandler(ctx, name, FromExternalServiceHandlerToHandler(sync))
}

func (c *externalServiceController) OnRemove(ctx context.Context, name string, sync ExternalServiceHandler) {
	removeHandler := generic.NewRemoveHandler(name, c.Updater(), FromExternalServiceHandlerToHandler(sync))
	c.AddGenericHandler(ctx, name, removeHandler)
}

func (c *externalServiceController) Enqueue(namespace, name string) {
	c.controllerManager.Enqueue(c.gvk, namespace, name)
}

func (c *externalServiceController) Informer() cache.SharedIndexInformer {
	return c.informer.Informer()
}

func (c *externalServiceController) GroupVersionKind() schema.GroupVersionKind {
	return c.gvk
}

func (c *externalServiceController) Cache() ExternalServiceCache {
	return &externalServiceCache{
		lister:  c.informer.Lister(),
		indexer: c.informer.Informer().GetIndexer(),
	}
}

func (c *externalServiceController) Create(obj *v1.ExternalService) (*v1.ExternalService, error) {
	return c.clientGetter.ExternalServices(obj.Namespace).Create(obj)
}

func (c *externalServiceController) Update(obj *v1.ExternalService) (*v1.ExternalService, error) {
	return c.clientGetter.ExternalServices(obj.Namespace).Update(obj)
}

func (c *externalServiceController) UpdateStatus(obj *v1.ExternalService) (*v1.ExternalService, error) {
	return c.clientGetter.ExternalServices(obj.Namespace).UpdateStatus(obj)
}

func (c *externalServiceController) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	return c.clientGetter.ExternalServices(namespace).Delete(name, options)
}

func (c *externalServiceController) Get(namespace, name string, options metav1.GetOptions) (*v1.ExternalService, error) {
	return c.clientGetter.ExternalServices(namespace).Get(name, options)
}

func (c *externalServiceController) List(namespace string, opts metav1.ListOptions) (*v1.ExternalServiceList, error) {
	return c.clientGetter.ExternalServices(namespace).List(opts)
}

func (c *externalServiceController) Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	return c.clientGetter.ExternalServices(namespace).Watch(opts)
}

func (c *externalServiceController) Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.ExternalService, err error) {
	return c.clientGetter.ExternalServices(namespace).Patch(name, pt, data, subresources...)
}

type externalServiceCache struct {
	lister  listers.ExternalServiceLister
	indexer cache.Indexer
}

func (c *externalServiceCache) Get(namespace, name string) (*v1.ExternalService, error) {
	return c.lister.ExternalServices(namespace).Get(name)
}

func (c *externalServiceCache) List(namespace string, selector labels.Selector) ([]*v1.ExternalService, error) {
	return c.lister.ExternalServices(namespace).List(selector)
}

func (c *externalServiceCache) AddIndexer(indexName string, indexer ExternalServiceIndexer) {
	utilruntime.Must(c.indexer.AddIndexers(map[string]cache.IndexFunc{
		indexName: func(obj interface{}) (strings []string, e error) {
			return indexer(obj.(*v1.ExternalService))
		},
	}))
}

func (c *externalServiceCache) GetByIndex(indexName, key string) (result []*v1.ExternalService, err error) {
	objs, err := c.indexer.ByIndex(indexName, key)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		result = append(result, obj.(*v1.ExternalService))
	}
	return result, nil
}
