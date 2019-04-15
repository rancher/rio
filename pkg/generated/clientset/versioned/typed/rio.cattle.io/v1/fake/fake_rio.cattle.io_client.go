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
	v1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/rio.cattle.io/v1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeRioV1 struct {
	*testing.Fake
}

func (c *FakeRioV1) ExternalServices(namespace string) v1.ExternalServiceInterface {
	return &FakeExternalServices{c, namespace}
}

func (c *FakeRioV1) PublicDomains(namespace string) v1.PublicDomainInterface {
	return &FakePublicDomains{c, namespace}
}

func (c *FakeRioV1) Routers(namespace string) v1.RouterInterface {
	return &FakeRouters{c, namespace}
}

func (c *FakeRioV1) Services(namespace string) v1.ServiceInterface {
	return &FakeServices{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeRioV1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
