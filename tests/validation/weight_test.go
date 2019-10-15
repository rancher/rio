package validation

import (
	"testing"
	"time"

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

	when("a staged service incrementally rolls out weight", func() {
		it("should slowly increase weight on the staged service and decrease current service weight", func() {
			// This test is known to fail until https://github.com/rancher/rio/issues/577 is resolved.
			// The time from rollout to obtaining the current weight, without Sleep, is 2 seconds.
			// Sleeping 8 seconds here with a rollout-interval of 4 seconds to guarantee 2 rollout ticks, plus the initial tick, with 2 seconds to spare.
			stagedService.Weight(80, true, 10, 4)
			time.Sleep(8 * time.Second)

			stagedWeightAfter10Seconds := stagedService.GetCurrentWeight()
			serviceWeightAfter10Seconds := service.GetCurrentWeight()
			assert.Equal(t, 30, stagedWeightAfter10Seconds)
			assert.Equal(t, 70, serviceWeightAfter10Seconds)
		})
	}, spec.Parallel())
}
