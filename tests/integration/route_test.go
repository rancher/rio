package integration

import (
	"testing"

	"github.com/rancher/rio/tests/testutil"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
)

func routeTests(t *testing.T, when spec.G, it spec.S) {

	var service testutil.TestService
	var stagedService testutil.TestService
	var routeA testutil.TestRoute
	var routeB testutil.TestRoute

	it.Before(func() {
		service.Create(t, "ibuildthecloud/demo:v1")
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
			routeA.Add(t, "/to-svc-v0", "to", service)
			routeB.Add(t, "/to-svc-v3", "to", stagedService)
			assert.Equal(t, "Hello World", routeA.GetEndpoint())
			assert.Equal(t, "Hello World v3", routeB.GetEndpoint())
		})
	}, spec.Parallel())
}
