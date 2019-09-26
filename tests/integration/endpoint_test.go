package integration

import (
	"testing"

	"github.com/rancher/rio/tests/testutil"
	"github.com/stretchr/testify/assert"

	"github.com/sclevine/spec"
)

func endpointTests(t *testing.T, when spec.G, it spec.S) {

	var service testutil.TestService
	var stagedService testutil.TestService

	it.After(func() {
		service.Remove()
		stagedService.Remove()
	})

	when("a service is running and another is staged", func() {
		it("should have endpoints that are available on both", func() {
			service.Create(t, "ibuildthecloud/demo:v1")
			stagedService = service.Stage("ibuildthecloud/demo:v3", "v3")
			assert.Equal(t, "Hello World", service.GetEndpoint())
			assert.Equal(t, "Hello World v3", stagedService.GetEndpoint())
		})
	}, spec.Parallel())
}
