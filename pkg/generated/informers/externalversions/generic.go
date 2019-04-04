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

package externalversions

import (
	"fmt"

	v1 "github.com/rancher/rio/pkg/apis/autoscale.rio.cattle.io/v1"
	projectriocattleiov1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	riocattleiov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	webhookinatorriocattleiov1 "github.com/rancher/rio/pkg/apis/webhookinator.rio.cattle.io/v1"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	cache "k8s.io/client-go/tools/cache"
)

// GenericInformer is type of SharedIndexInformer which will locate and delegate to other
// sharedInformers based on type
type GenericInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() cache.GenericLister
}

type genericInformer struct {
	informer cache.SharedIndexInformer
	resource schema.GroupResource
}

// Informer returns the SharedIndexInformer.
func (f *genericInformer) Informer() cache.SharedIndexInformer {
	return f.informer
}

// Lister returns the GenericLister.
func (f *genericInformer) Lister() cache.GenericLister {
	return cache.NewGenericLister(f.Informer().GetIndexer(), f.resource)
}

// ForResource gives generic access to a shared informer of the matching type
// TODO extend this to unknown resources with a client pool
func (f *sharedInformerFactory) ForResource(resource schema.GroupVersionResource) (GenericInformer, error) {
	switch resource {
	// Group=autoscale.rio.cattle.io, Version=v1
	case v1.SchemeGroupVersion.WithResource("servicescalerecommendations"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscale().V1().ServiceScaleRecommendations().Informer()}, nil

		// Group=project.rio.cattle.io, Version=v1
	case projectriocattleiov1.SchemeGroupVersion.WithResource("clusterdomains"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Project().V1().ClusterDomains().Informer()}, nil
	case projectriocattleiov1.SchemeGroupVersion.WithResource("features"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Project().V1().Features().Informer()}, nil
	case projectriocattleiov1.SchemeGroupVersion.WithResource("publicdomains"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Project().V1().PublicDomains().Informer()}, nil

		// Group=rio.cattle.io, Version=v1
	case riocattleiov1.SchemeGroupVersion.WithResource("externalservices"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Rio().V1().ExternalServices().Informer()}, nil
	case riocattleiov1.SchemeGroupVersion.WithResource("routers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Rio().V1().Routers().Informer()}, nil
	case riocattleiov1.SchemeGroupVersion.WithResource("services"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Rio().V1().Services().Informer()}, nil

		// Group=webhookinator.rio.cattle.io, Version=v1
	case webhookinatorriocattleiov1.SchemeGroupVersion.WithResource("gitwebhookexecutions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Webhookinator().V1().GitWebHookExecutions().Informer()}, nil
	case webhookinatorriocattleiov1.SchemeGroupVersion.WithResource("gitwebhookreceivers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Webhookinator().V1().GitWebHookReceivers().Informer()}, nil

	}

	return nil, fmt.Errorf("no informer found for %v", resource)
}
