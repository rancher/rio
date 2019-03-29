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
	projectriocattleiov1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeClusterDomains implements ClusterDomainInterface
type FakeClusterDomains struct {
	Fake *FakeProjectV1
	ns   string
}

var clusterdomainsResource = schema.GroupVersionResource{Group: "project.rio.cattle.io", Version: "v1", Resource: "clusterdomains"}

var clusterdomainsKind = schema.GroupVersionKind{Group: "project.rio.cattle.io", Version: "v1", Kind: "ClusterDomain"}

// Get takes name of the clusterDomain, and returns the corresponding clusterDomain object, and an error if there is any.
func (c *FakeClusterDomains) Get(name string, options v1.GetOptions) (result *projectriocattleiov1.ClusterDomain, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(clusterdomainsResource, c.ns, name), &projectriocattleiov1.ClusterDomain{})

	if obj == nil {
		return nil, err
	}
	return obj.(*projectriocattleiov1.ClusterDomain), err
}

// List takes label and field selectors, and returns the list of ClusterDomains that match those selectors.
func (c *FakeClusterDomains) List(opts v1.ListOptions) (result *projectriocattleiov1.ClusterDomainList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(clusterdomainsResource, clusterdomainsKind, c.ns, opts), &projectriocattleiov1.ClusterDomainList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &projectriocattleiov1.ClusterDomainList{ListMeta: obj.(*projectriocattleiov1.ClusterDomainList).ListMeta}
	for _, item := range obj.(*projectriocattleiov1.ClusterDomainList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested clusterDomains.
func (c *FakeClusterDomains) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(clusterdomainsResource, c.ns, opts))

}

// Create takes the representation of a clusterDomain and creates it.  Returns the server's representation of the clusterDomain, and an error, if there is any.
func (c *FakeClusterDomains) Create(clusterDomain *projectriocattleiov1.ClusterDomain) (result *projectriocattleiov1.ClusterDomain, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(clusterdomainsResource, c.ns, clusterDomain), &projectriocattleiov1.ClusterDomain{})

	if obj == nil {
		return nil, err
	}
	return obj.(*projectriocattleiov1.ClusterDomain), err
}

// Update takes the representation of a clusterDomain and updates it. Returns the server's representation of the clusterDomain, and an error, if there is any.
func (c *FakeClusterDomains) Update(clusterDomain *projectriocattleiov1.ClusterDomain) (result *projectriocattleiov1.ClusterDomain, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(clusterdomainsResource, c.ns, clusterDomain), &projectriocattleiov1.ClusterDomain{})

	if obj == nil {
		return nil, err
	}
	return obj.(*projectriocattleiov1.ClusterDomain), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeClusterDomains) UpdateStatus(clusterDomain *projectriocattleiov1.ClusterDomain) (*projectriocattleiov1.ClusterDomain, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(clusterdomainsResource, "status", c.ns, clusterDomain), &projectriocattleiov1.ClusterDomain{})

	if obj == nil {
		return nil, err
	}
	return obj.(*projectriocattleiov1.ClusterDomain), err
}

// Delete takes name of the clusterDomain and deletes it. Returns an error if one occurs.
func (c *FakeClusterDomains) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(clusterdomainsResource, c.ns, name), &projectriocattleiov1.ClusterDomain{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeClusterDomains) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(clusterdomainsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &projectriocattleiov1.ClusterDomainList{})
	return err
}

// Patch applies the patch and returns the patched clusterDomain.
func (c *FakeClusterDomains) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *projectriocattleiov1.ClusterDomain, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(clusterdomainsResource, c.ns, name, pt, data, subresources...), &projectriocattleiov1.ClusterDomain{})

	if obj == nil {
		return nil, err
	}
	return obj.(*projectriocattleiov1.ClusterDomain), err
}
