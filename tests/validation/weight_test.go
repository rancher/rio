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

	when("a staged service incrementally rolls out weight", func() {
		it.Before(func() {
			service.Create(t, "--weight", "100", "ibuildthecloud/demo:v1")
			stagedService = service.Stage("ibuildthecloud/demo:v3", "v3")
		})

		it.After(func() {
			service.Remove()
			stagedService.Remove()
		})
		it("should slowly increase weight on the staged service and leave service weight unchanged", func() {
			assert.Equal(t, 100, service.GetComputedWeight())
			// The time from rollout to obtaining the current weight, without Sleep, is 2 seconds.
			// Sleeping 9 seconds to guarantee 5 rollout ticks with 1 second to spare since the default tick interval is 2 seconds.
			stagedService.WeightWithoutWaiting(60, "--duration=1m")
			time.Sleep(8 * time.Second)
			stagedComputedWeightAfter10Seconds, serviceComputedWeightAfter10Seconds := stagedService.GetComputedWeight(), service.GetComputedWeight()
			assert.Less(t, 10, stagedComputedWeightAfter10Seconds)
			assert.Greater(t, 30, stagedComputedWeightAfter10Seconds)
			assert.Equal(t, 100, serviceComputedWeightAfter10Seconds)
		})
	}, spec.Parallel())
}
