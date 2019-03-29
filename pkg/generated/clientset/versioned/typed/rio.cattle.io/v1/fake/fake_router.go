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
	riocattleiov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeRouters implements RouterInterface
type FakeRouters struct {
	Fake *FakeRioV1
	ns   string
}

var routersResource = schema.GroupVersionResource{Group: "rio.cattle.io", Version: "v1", Resource: "routers"}

var routersKind = schema.GroupVersionKind{Group: "rio.cattle.io", Version: "v1", Kind: "Router"}

// Get takes name of the router, and returns the corresponding router object, and an error if there is any.
func (c *FakeRouters) Get(name string, options v1.GetOptions) (result *riocattleiov1.Router, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(routersResource, c.ns, name), &riocattleiov1.Router{})

	if obj == nil {
		return nil, err
	}
	return obj.(*riocattleiov1.Router), err
}

// List takes label and field selectors, and returns the list of Routers that match those selectors.
func (c *FakeRouters) List(opts v1.ListOptions) (result *riocattleiov1.RouterList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(routersResource, routersKind, c.ns, opts), &riocattleiov1.RouterList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &riocattleiov1.RouterList{ListMeta: obj.(*riocattleiov1.RouterList).ListMeta}
	for _, item := range obj.(*riocattleiov1.RouterList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested routers.
func (c *FakeRouters) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(routersResource, c.ns, opts))

}

// Create takes the representation of a router and creates it.  Returns the server's representation of the router, and an error, if there is any.
func (c *FakeRouters) Create(router *riocattleiov1.Router) (result *riocattleiov1.Router, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(routersResource, c.ns, router), &riocattleiov1.Router{})

	if obj == nil {
		return nil, err
	}
	return obj.(*riocattleiov1.Router), err
}

// Update takes the representation of a router and updates it. Returns the server's representation of the router, and an error, if there is any.
func (c *FakeRouters) Update(router *riocattleiov1.Router) (result *riocattleiov1.Router, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(routersResource, c.ns, router), &riocattleiov1.Router{})

	if obj == nil {
		return nil, err
	}
	return obj.(*riocattleiov1.Router), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeRouters) UpdateStatus(router *riocattleiov1.Router) (*riocattleiov1.Router, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(routersResource, "status", c.ns, router), &riocattleiov1.Router{})

	if obj == nil {
		return nil, err
	}
	return obj.(*riocattleiov1.Router), err
}

// Delete takes name of the router and deletes it. Returns an error if one occurs.
func (c *FakeRouters) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(routersResource, c.ns, name), &riocattleiov1.Router{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeRouters) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(routersResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &riocattleiov1.RouterList{})
	return err
}

// Patch applies the patch and returns the patched router.
func (c *FakeRouters) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *riocattleiov1.Router, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(routersResource, c.ns, name, pt, data, subresources...), &riocattleiov1.Router{})

	if obj == nil {
		return nil, err
	}
	return obj.(*riocattleiov1.Router), err
}
