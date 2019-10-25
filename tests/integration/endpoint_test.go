package integration

import (
	"github.com/rancher/rio/tests/testutil"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
	"testing"
)

func endpointTests(t *testing.T, when spec.G, it spec.S) {

	var service testutil.TestService
	var stagedService testutil.TestService

	when("a service is running and another is staged", func() {

		it.Before(func() {
			service.Create(t, "ibuildthecloud/demo:v1")
		})
		it.After(func() {
			service.Remove()
			stagedService.Remove()
		})

		it("should have endpoints that are available on both with one app endpoint", func() {
			stagedService = service.Stage("ibuildthecloud/demo:v3", "v3")

			// todo: fix kubectl func
			// Check the hostnames returned by Rio and Kubectl are equal
			//assert.Equal(t,
			//	testutil.GetHostname(service.GetKubeEndpointURLs()),
			//	testutil.GetHostname(service.GetEndpointURL()),
			//)
			//assert.Equal(t,
			//	testutil.GetHostname(stagedService.GetKubeEndpointURLs()),
			//	testutil.GetHostname(stagedService.GetEndpointURL()),
			//)

			assert.Equal(t, "Hello World", service.GetEndpointResponse())
			assert.Equal(t, "Hello World v3", stagedService.GetEndpointResponse())

			// todo: fix kubectl func
			//assert.Equal(t,
			//	testutil.GetHostname(service.GetKubeAppEndpointURLs()[0]),
			//	testutil.GetHostname(service.GetAppEndpointURLs()[0]),
			//)
			assert.Equal(t,
				testutil.GetHostname(service.GetAppEndpointURLs()[0]),
				testutil.GetHostname(stagedService.GetAppEndpointURLs()[0]),
			)
			assert.Equal(t, "Hello World", service.GetAppEndpointResponse())
		})
	}, spec.Parallel())

	when("a staged service is promoted", func() {
		it.Before(func() {
			service.Create(t, "ibuildthecloud/demo:v1")
			stagedService = service.Stage("ibuildthecloud/demo:v3", "v3")
			stagedService.Promote()
		})
		it.After(func() {
			service.Remove()
			stagedService.Remove()
		})

		it("should retain all revision endpoints with an app endpoint pointing to the new revision", func() {
			assert.Equal(t, "Hello World", service.GetEndpointResponse())
			assert.Equal(t, "Hello World v3", stagedService.GetEndpointResponse())
			// todo: fix kubectl func
			//assert.Equal(t, 0, service.GetKubeCurrentWeight())
			//assert.Equal(t, 100, stagedService.GetKubeCurrentWeight())
			assert.Equal(t, "Hello World v3", service.GetAppEndpointResponse())
		})
		it("should allow rolling back to the previous revision", func() {
			service.Promote()
			// todo: fix kubectl func
			//assert.Equal(t, 100, service.GetKubeCurrentWeight())
			//assert.Equal(t, 0, stagedService.GetKubeCurrentWeight())
			assert.Equal(t, "Hello World", service.GetEndpointResponse())
		})
	}, spec.Parallel())
}
