package integration

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"

	"github.com/rancher/rio/tests/testutil"
)

func endpointTests(t *testing.T, when spec.G, it spec.S) {

	var service testutil.TestService
	var stagedService testutil.TestService

	it.Before(func() {
		service.Create(t, "ibuildthecloud/demo:v1")
	})
	it.After(func() {
		service.Remove()
		stagedService.Remove()
	})

	when("a service is running and another is staged", func() {
		it("should have endpoints that are available on both", func() {
			stagedService = service.Stage("ibuildthecloud/demo:v3", "v3")

			// Check the hostnames returned by Rio and Kubectl are equal
			assert.Equal(t,
				testutil.GetHostname(service.GetKubeEndpointURL()),
				testutil.GetHostname(service.GetEndpointURL()),
			)
			assert.Equal(t,
				testutil.GetHostname(stagedService.GetKubeEndpointURL()),
				testutil.GetHostname(stagedService.GetEndpointURL()),
			)

			assert.Equal(t, "Hello World", service.GetEndpointResponse())
			assert.Equal(t, "Hello World v3", stagedService.GetEndpointResponse())
		})
		it("should have the service app endpoint properly created", func() {

			// Check the hostnames returned by Rio and Kubectl are equal
			assert.Equal(t,
				testutil.GetHostname(service.GetKubeAppEndpointURL()),
				testutil.GetHostname(service.GetAppEndpointURL()),
			)
			assert.Equal(t, "Hello World", service.GetAppEndpointResponse())
		})
	}, spec.Parallel())
}
