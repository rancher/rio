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

	v1 "github.com/rancher/rio/pkg/apis/autoscale.rio.cattle.io/v1"
	clientset "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/autoscale.rio.cattle.io/v1"
	informers "github.com/rancher/rio/pkg/generated/informers/externalversions/autoscale.rio.cattle.io/v1"
	listers "github.com/rancher/rio/pkg/generated/listers/autoscale.rio.cattle.io/v1"
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

type ServiceScaleRecommendationHandler func(string, *v1.ServiceScaleRecommendation) (*v1.ServiceScaleRecommendation, error)

type ServiceScaleRecommendationController interface {
	ServiceScaleRecommendationClient

	OnChange(ctx context.Context, name string, sync ServiceScaleRecommendationHandler)
	OnRemove(ctx context.Context, name string, sync ServiceScaleRecommendationHandler)
	Enqueue(namespace, name string)

	Cache() ServiceScaleRecommendationCache

	Informer() cache.SharedIndexInformer
	GroupVersionKind() schema.GroupVersionKind

	AddGenericHandler(ctx context.Context, name string, handler generic.Handler)
	AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler)
	Updater() generic.Updater
}

type ServiceScaleRecommendationClient interface {
	Create(*v1.ServiceScaleRecommendation) (*v1.ServiceScaleRecommendation, error)
	Update(*v1.ServiceScaleRecommendation) (*v1.ServiceScaleRecommendation, error)
	UpdateStatus(*v1.ServiceScaleRecommendation) (*v1.ServiceScaleRecommendation, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	Get(namespace, name string, options metav1.GetOptions) (*v1.ServiceScaleRecommendation, error)
	List(namespace string, opts metav1.ListOptions) (*v1.ServiceScaleRecommendationList, error)
	Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error)
	Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.ServiceScaleRecommendation, err error)
}

type ServiceScaleRecommendationCache interface {
	Get(namespace, name string) (*v1.ServiceScaleRecommendation, error)
	List(namespace string, selector labels.Selector) ([]*v1.ServiceScaleRecommendation, error)

	AddIndexer(indexName string, indexer ServiceScaleRecommendationIndexer)
	GetByIndex(indexName, key string) ([]*v1.ServiceScaleRecommendation, error)
}

type ServiceScaleRecommendationIndexer func(obj *v1.ServiceScaleRecommendation) ([]string, error)

type serviceScaleRecommendationController struct {
	controllerManager *generic.ControllerManager
	clientGetter      clientset.ServiceScaleRecommendationsGetter
	informer          informers.ServiceScaleRecommendationInformer
	gvk               schema.GroupVersionKind
}

func NewServiceScaleRecommendationController(gvk schema.GroupVersionKind, controllerManager *generic.ControllerManager, clientGetter clientset.ServiceScaleRecommendationsGetter, informer informers.ServiceScaleRecommendationInformer) ServiceScaleRecommendationController {
	return &serviceScaleRecommendationController{
		controllerManager: controllerManager,
		clientGetter:      clientGetter,
		informer:          informer,
		gvk:               gvk,
	}
}

func FromServiceScaleRecommendationHandlerToHandler(sync ServiceScaleRecommendationHandler) generic.Handler {
	return func(key string, obj runtime.Object) (ret runtime.Object, err error) {
		var v *v1.ServiceScaleRecommendation
		if obj == nil {
			v, err = sync(key, nil)
		} else {
			v, err = sync(key, obj.(*v1.ServiceScaleRecommendation))
		}
		if v == nil {
			return nil, err
		}
		return v, err
	}
}

func (c *serviceScaleRecommendationController) Updater() generic.Updater {
	return func(obj runtime.Object) (runtime.Object, error) {
		newObj, err := c.Update(obj.(*v1.ServiceScaleRecommendation))
		if newObj == nil {
			return nil, err
		}
		return newObj, err
	}
}

func UpdateServiceScaleRecommendationOnChange(updater generic.Updater, handler ServiceScaleRecommendationHandler) ServiceScaleRecommendationHandler {
	return func(key string, obj *v1.ServiceScaleRecommendation) (*v1.ServiceScaleRecommendation, error) {
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
				copyObj = newObj.(*v1.ServiceScaleRecommendation)
			}
		}

		return copyObj, err
	}
}

func (c *serviceScaleRecommendationController) AddGenericHandler(ctx context.Context, name string, handler generic.Handler) {
	c.controllerManager.AddHandler(ctx, c.gvk, c.informer.Informer(), name, handler)
}

func (c *serviceScaleRecommendationController) AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler) {
	removeHandler := generic.NewRemoveHandler(name, c.Updater(), handler)
	c.controllerManager.AddHandler(ctx, c.gvk, c.informer.Informer(), name, removeHandler)
}

func (c *serviceScaleRecommendationController) OnChange(ctx context.Context, name string, sync ServiceScaleRecommendationHandler) {
	c.AddGenericHandler(ctx, name, FromServiceScaleRecommendationHandlerToHandler(sync))
}

func (c *serviceScaleRecommendationController) OnRemove(ctx context.Context, name string, sync ServiceScaleRecommendationHandler) {
	removeHandler := generic.NewRemoveHandler(name, c.Updater(), FromServiceScaleRecommendationHandlerToHandler(sync))
	c.AddGenericHandler(ctx, name, removeHandler)
}

func (c *serviceScaleRecommendationController) Enqueue(namespace, name string) {
	c.controllerManager.Enqueue(c.gvk, namespace, name)
}

func (c *serviceScaleRecommendationController) Informer() cache.SharedIndexInformer {
	return c.informer.Informer()
}

func (c *serviceScaleRecommendationController) GroupVersionKind() schema.GroupVersionKind {
	return c.gvk
}

func (c *serviceScaleRecommendationController) Cache() ServiceScaleRecommendationCache {
	return &serviceScaleRecommendationCache{
		lister:  c.informer.Lister(),
		indexer: c.informer.Informer().GetIndexer(),
	}
}

func (c *serviceScaleRecommendationController) Create(obj *v1.ServiceScaleRecommendation) (*v1.ServiceScaleRecommendation, error) {
	return c.clientGetter.ServiceScaleRecommendations(obj.Namespace).Create(obj)
}

func (c *serviceScaleRecommendationController) Update(obj *v1.ServiceScaleRecommendation) (*v1.ServiceScaleRecommendation, error) {
	return c.clientGetter.ServiceScaleRecommendations(obj.Namespace).Update(obj)
}

func (c *serviceScaleRecommendationController) UpdateStatus(obj *v1.ServiceScaleRecommendation) (*v1.ServiceScaleRecommendation, error) {
	return c.clientGetter.ServiceScaleRecommendations(obj.Namespace).UpdateStatus(obj)
}

func (c *serviceScaleRecommendationController) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	return c.clientGetter.ServiceScaleRecommendations(namespace).Delete(name, options)
}

func (c *serviceScaleRecommendationController) Get(namespace, name string, options metav1.GetOptions) (*v1.ServiceScaleRecommendation, error) {
	return c.clientGetter.ServiceScaleRecommendations(namespace).Get(name, options)
}

func (c *serviceScaleRecommendationController) List(namespace string, opts metav1.ListOptions) (*v1.ServiceScaleRecommendationList, error) {
	return c.clientGetter.ServiceScaleRecommendations(namespace).List(opts)
}

func (c *serviceScaleRecommendationController) Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	return c.clientGetter.ServiceScaleRecommendations(namespace).Watch(opts)
}

func (c *serviceScaleRecommendationController) Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.ServiceScaleRecommendation, err error) {
	return c.clientGetter.ServiceScaleRecommendations(namespace).Patch(name, pt, data, subresources...)
}

type serviceScaleRecommendationCache struct {
	lister  listers.ServiceScaleRecommendationLister
	indexer cache.Indexer
}

func (c *serviceScaleRecommendationCache) Get(namespace, name string) (*v1.ServiceScaleRecommendation, error) {
	return c.lister.ServiceScaleRecommendations(namespace).Get(name)
}

func (c *serviceScaleRecommendationCache) List(namespace string, selector labels.Selector) ([]*v1.ServiceScaleRecommendation, error) {
	return c.lister.ServiceScaleRecommendations(namespace).List(selector)
}

func (c *serviceScaleRecommendationCache) AddIndexer(indexName string, indexer ServiceScaleRecommendationIndexer) {
	utilruntime.Must(c.indexer.AddIndexers(map[string]cache.IndexFunc{
		indexName: func(obj interface{}) (strings []string, e error) {
			return indexer(obj.(*v1.ServiceScaleRecommendation))
		},
	}))
}

func (c *serviceScaleRecommendationCache) GetByIndex(indexName, key string) (result []*v1.ServiceScaleRecommendation, err error) {
	objs, err := c.indexer.ByIndex(indexName, key)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		result = append(result, obj.(*v1.ServiceScaleRecommendation))
	}
	return result, nil
}
