package integration

import (
	"testing"

	"github.com/rancher/rio/tests/testutil"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
)

func weightTests(t *testing.T, when spec.G, it spec.S) {

	var service testutil.TestService
	var stagedService testutil.TestService

	it.After(func() {
		service.Remove()
		stagedService.Remove()
	})

	when("a running service has a new imaged staged", func() {
		it.Before(func() {
			service.Create(t, "--weight", "100", "ibuildthecloud/demo:v1")
			stagedService = service.Stage("ibuildthecloud/demo:v3", "v3")
		})

		it("should be able to split weights between revisions", func() {
			stagedService.Weight(40, "--duration=0s")
			assert.Equal(t, 40, stagedService.GetCurrentWeight())
			assert.Equal(t, 60, service.GetCurrentWeight())
			responses := service.GetResponseCounts([]string{"Hello World", "Hello World v3"}, 12)
			assert.Greater(t, responses["Hello World"], 2, "The application did not return enough responses from the service. which has slightly more weight than the staged service.")
			assert.GreaterOrEqual(t, responses["Hello World v3"], 1, "The application did not return enough responses from the staged service. which has slightly less weight than the service.")
		})
	}, spec.Parallel())

	when("a staged service is promoted", func() {
		it.Before(func() {
			service.Create(t, "--weight", "100", "ibuildthecloud/demo:v1")
			stagedService = service.Stage("ibuildthecloud/demo:v3", "v3")
			stagedService.Promote()
		})

		it("should retain all revision endpoints with an app endpoint pointing to the new version", func() {
			assert.Equal(t, "Hello World", service.GetEndpointResponse())
			assert.Equal(t, "Hello World v3", stagedService.GetEndpointResponse())
			assert.Equal(t, 0, service.GetCurrentWeight())
			assert.Equal(t, 100, stagedService.GetCurrentWeight())
			assert.Equal(t, "Hello World v3", service.GetAppEndpointResponse())
		})
		it("should allow rolling back to the previous version", func() {
			service.Promote()
			assert.Equal(t, 100, service.GetCurrentWeight())
			assert.Equal(t, 0, stagedService.GetCurrentWeight())
			assert.Equal(t, "Hello World", service.GetAppEndpointResponse())
		})
	}, spec.Parallel())
}
