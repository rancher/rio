package integration

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"

	"github.com/rancher/rio/tests/testutil"
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
		})
		it("should be able to move 5% of weight to new revision", func() {
			stagedService.Weight(5)
			assert.Equal(t, 95, service.GetCurrentWeight())
			assert.Equal(t, 5, stagedService.GetCurrentWeight())
		})
	}, spec.Parallel())
}
