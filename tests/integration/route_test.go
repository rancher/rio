package integration

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"

	"github.com/rancher/rio/tests/testutil"
)

func routeTests(t *testing.T, when spec.G, it spec.S) {

	var service testutil.TestService
	var stagedService testutil.TestService
	var routeA testutil.TestRoute
	var routeB testutil.TestRoute

	it.Before(func() {
		service.Create(t, "--weight", "100", "ibuildthecloud/demo:v1")
		stagedService = service.Stage("ibuildthecloud/demo:v3", "v3")
	})

	it.After(func() {
		service.Remove()
		stagedService.Remove()
		routeA.Remove()
		routeB.Remove()
	})

	when("a running service has routes added to it", func() {
		it("should be accessible on multiple revision by multiple routes", func() {
			routeA.Add(t, "", "/to-svc-v0", "to", service)
			routeB.Add(t, "", "/to-svc-v3", "to", stagedService)
			assert.Equal(t, "Hello World", routeA.GetEndpointResponse())
			assert.Equal(t, "Hello World v3", routeB.GetEndpointResponse())
			assert.Equal(t, "Hello World", routeA.GetKubeEndpointResponse())
			assert.Equal(t, "Hello World v3", routeB.GetKubeEndpointResponse())
		})
		it("should be accessible from one domain", func() {
			routeA.Add(t, "test-route-root", "/first", "to", service)
			routeB.Add(t, "test-route-root", "/to-svc-v3", "to", stagedService)
			assert.Equal(t, routeA.Router.Status.Endpoints, routeB.Router.Status.Endpoints)
			assert.Equal(t, routeA.Name, routeB.Name)
			assert.Equal(t, "Hello World", routeA.GetEndpointResponse())
			assert.Equal(t, "Hello World v3", routeB.GetEndpointResponse())
			assert.Equal(t, "Hello World", routeA.GetKubeEndpointResponse())
			assert.Equal(t, "Hello World v3", routeB.GetKubeEndpointResponse())
		})
	}, spec.Parallel())
}
