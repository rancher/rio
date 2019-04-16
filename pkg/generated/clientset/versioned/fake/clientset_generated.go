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

package fake

import (
	clientset "github.com/rancher/rio/pkg/generated/clientset/versioned"
	autoscalev1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/autoscale.rio.cattle.io/v1"
	fakeautoscalev1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/autoscale.rio.cattle.io/v1/fake"
	projectv1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/project.rio.cattle.io/v1"
	fakeprojectv1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/project.rio.cattle.io/v1/fake"
	riov1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/rio.cattle.io/v1"
	fakeriov1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/rio.cattle.io/v1/fake"
	webhookinatorv1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/webhookinator.rio.cattle.io/v1"
	fakewebhookinatorv1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/webhookinator.rio.cattle.io/v1/fake"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/testing"
)

// NewSimpleClientset returns a clientset that will respond with the provided objects.
// It's backed by a very simple object tracker that processes creates, updates and deletions as-is,
// without applying any validations and/or defaults. It shouldn't be considered a replacement
// for a real clientset and is mostly useful in simple unit tests.
func NewSimpleClientset(objects ...runtime.Object) *Clientset {
	o := testing.NewObjectTracker(scheme, codecs.UniversalDecoder())
	for _, obj := range objects {
		if err := o.Add(obj); err != nil {
			panic(err)
		}
	}

	cs := &Clientset{}
	cs.discovery = &fakediscovery.FakeDiscovery{Fake: &cs.Fake}
	cs.AddReactor("*", "*", testing.ObjectReaction(o))
	cs.AddWatchReactor("*", func(action testing.Action) (handled bool, ret watch.Interface, err error) {
		gvr := action.GetResource()
		ns := action.GetNamespace()
		watch, err := o.Watch(gvr, ns)
		if err != nil {
			return false, nil, err
		}
		return true, watch, nil
	})

	return cs
}

// Clientset implements clientset.Interface. Meant to be embedded into a
// struct to get a default implementation. This makes faking out just the method
// you want to test easier.
type Clientset struct {
	testing.Fake
	discovery *fakediscovery.FakeDiscovery
}

func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	return c.discovery
}

var _ clientset.Interface = &Clientset{}

// RioV1 retrieves the RioV1Client
func (c *Clientset) RioV1() riov1.RioV1Interface {
	return &fakeriov1.FakeRioV1{Fake: &c.Fake}
}

// Rio retrieves the RioV1Client
func (c *Clientset) Rio() riov1.RioV1Interface {
	return &fakeriov1.FakeRioV1{Fake: &c.Fake}
}

// WebhookinatorV1 retrieves the WebhookinatorV1Client
func (c *Clientset) WebhookinatorV1() webhookinatorv1.WebhookinatorV1Interface {
	return &fakewebhookinatorv1.FakeWebhookinatorV1{Fake: &c.Fake}
}

// Webhookinator retrieves the WebhookinatorV1Client
func (c *Clientset) Webhookinator() webhookinatorv1.WebhookinatorV1Interface {
	return &fakewebhookinatorv1.FakeWebhookinatorV1{Fake: &c.Fake}
}

// ProjectV1 retrieves the ProjectV1Client
func (c *Clientset) ProjectV1() projectv1.ProjectV1Interface {
	return &fakeprojectv1.FakeProjectV1{Fake: &c.Fake}
}

// Project retrieves the ProjectV1Client
func (c *Clientset) Project() projectv1.ProjectV1Interface {
	return &fakeprojectv1.FakeProjectV1{Fake: &c.Fake}
}

// AutoscaleV1 retrieves the AutoscaleV1Client
func (c *Clientset) AutoscaleV1() autoscalev1.AutoscaleV1Interface {
	return &fakeautoscalev1.FakeAutoscaleV1{Fake: &c.Fake}
}

// Autoscale retrieves the AutoscaleV1Client
func (c *Clientset) Autoscale() autoscalev1.AutoscaleV1Interface {
	return &fakeautoscalev1.FakeAutoscaleV1{Fake: &c.Fake}
}
