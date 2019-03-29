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

// FakePublicDomains implements PublicDomainInterface
type FakePublicDomains struct {
	Fake *FakeProjectV1
	ns   string
}

var publicdomainsResource = schema.GroupVersionResource{Group: "project.rio.cattle.io", Version: "v1", Resource: "publicdomains"}

var publicdomainsKind = schema.GroupVersionKind{Group: "project.rio.cattle.io", Version: "v1", Kind: "PublicDomain"}

// Get takes name of the publicDomain, and returns the corresponding publicDomain object, and an error if there is any.
func (c *FakePublicDomains) Get(name string, options v1.GetOptions) (result *projectriocattleiov1.PublicDomain, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(publicdomainsResource, c.ns, name), &projectriocattleiov1.PublicDomain{})

	if obj == nil {
		return nil, err
	}
	return obj.(*projectriocattleiov1.PublicDomain), err
}

// List takes label and field selectors, and returns the list of PublicDomains that match those selectors.
func (c *FakePublicDomains) List(opts v1.ListOptions) (result *projectriocattleiov1.PublicDomainList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(publicdomainsResource, publicdomainsKind, c.ns, opts), &projectriocattleiov1.PublicDomainList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &projectriocattleiov1.PublicDomainList{ListMeta: obj.(*projectriocattleiov1.PublicDomainList).ListMeta}
	for _, item := range obj.(*projectriocattleiov1.PublicDomainList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested publicDomains.
func (c *FakePublicDomains) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(publicdomainsResource, c.ns, opts))

}

// Create takes the representation of a publicDomain and creates it.  Returns the server's representation of the publicDomain, and an error, if there is any.
func (c *FakePublicDomains) Create(publicDomain *projectriocattleiov1.PublicDomain) (result *projectriocattleiov1.PublicDomain, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(publicdomainsResource, c.ns, publicDomain), &projectriocattleiov1.PublicDomain{})

	if obj == nil {
		return nil, err
	}
	return obj.(*projectriocattleiov1.PublicDomain), err
}

// Update takes the representation of a publicDomain and updates it. Returns the server's representation of the publicDomain, and an error, if there is any.
func (c *FakePublicDomains) Update(publicDomain *projectriocattleiov1.PublicDomain) (result *projectriocattleiov1.PublicDomain, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(publicdomainsResource, c.ns, publicDomain), &projectriocattleiov1.PublicDomain{})

	if obj == nil {
		return nil, err
	}
	return obj.(*projectriocattleiov1.PublicDomain), err
}

// Delete takes name of the publicDomain and deletes it. Returns an error if one occurs.
func (c *FakePublicDomains) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(publicdomainsResource, c.ns, name), &projectriocattleiov1.PublicDomain{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePublicDomains) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(publicdomainsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &projectriocattleiov1.PublicDomainList{})
	return err
}

// Patch applies the patch and returns the patched publicDomain.
func (c *FakePublicDomains) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *projectriocattleiov1.PublicDomain, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(publicdomainsResource, c.ns, name, pt, data, subresources...), &projectriocattleiov1.PublicDomain{})

	if obj == nil {
		return nil, err
	}
	return obj.(*projectriocattleiov1.PublicDomain), err
}
