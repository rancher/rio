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
	internalinterfaces "github.com/rancher/rio/pkg/generated/informers/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// Configs returns a ConfigInformer.
	Configs() ConfigInformer
	// ExternalServices returns a ExternalServiceInformer.
	ExternalServices() ExternalServiceInformer
	// Routers returns a RouterInformer.
	Routers() RouterInformer
	// Services returns a ServiceInformer.
	Services() ServiceInformer
	// Stacks returns a StackInformer.
	Stacks() StackInformer
	// Volumes returns a VolumeInformer.
	Volumes() VolumeInformer
}

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// Configs returns a ConfigInformer.
func (v *version) Configs() ConfigInformer {
	return &configInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// ExternalServices returns a ExternalServiceInformer.
func (v *version) ExternalServices() ExternalServiceInformer {
	return &externalServiceInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Routers returns a RouterInformer.
func (v *version) Routers() RouterInformer {
	return &routerInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Services returns a ServiceInformer.
func (v *version) Services() ServiceInformer {
	return &serviceInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Stacks returns a StackInformer.
func (v *version) Stacks() StackInformer {
	return &stackInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Volumes returns a VolumeInformer.
func (v *version) Volumes() VolumeInformer {
	return &volumeInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}
