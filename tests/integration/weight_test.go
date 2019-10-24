package integration

import (
	"github.com/rancher/rio/tests/testutil"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
	"testing"
)

func weightTests(t *testing.T, when spec.G, it spec.S) {

	var service testutil.TestService
	var stagedService testutil.TestService

	it.Before(func() {
		service.Create(t, "ibuildthecloud/demo:v1")
		stagedService = service.Stage("ibuildthecloud/demo:v3", "v3")
	})

	it.After(func() {
		service.Remove()
		stagedService.Remove()
	})

	when("a running service has a new imaged staged", func() {
		it("should keep 100% of weight on original revision", func() {
			assert.Equal(t, 100, service.GetCurrentWeight())
			assert.Equal(t, 0, stagedService.GetCurrentWeight())
			//assert.Equal(t, 100, service.GetKubeCurrentWeight())
			//assert.Equal(t, 0, stagedService.GetKubeCurrentWeight())
		})
		it("should be able to split weights between revisions", func() {
			stagedService.Weight(40, false, 5, 5)
			assert.Equal(t, 60, service.GetCurrentWeight())
			assert.Equal(t, 40, stagedService.GetCurrentWeight())
			//assert.Equal(t, 60, service.GetKubeCurrentWeight())
			//assert.Equal(t, 40, stagedService.GetKubeCurrentWeight())
			responses := service.GetResponseCounts([]string{"Hello World", "Hello World v3"}, 12)
			assert.Greater(t, responses["Hello World"], 2, "The application did not return enough responses from the service. which has slightly more weight than the staged service.")
			assert.GreaterOrEqual(t, responses["Hello World v3"], 1, "The application did not return enough responses from the staged service. which has slightly less weight than the service.")
		})
	}, spec.Parallel())
}
