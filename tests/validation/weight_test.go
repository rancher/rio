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
			assert.Equal(t, 10000, service.GetComputedWeight())
			// The time from starting rollout to obtaining the current weight, without Sleep, is 1 to 2 seconds due to program execution time.
			// Sleeping 5 seconds to guarantee 2 rollout ticks with 1 to 2 seconds to spare since the default tick interval is 4 seconds.
			stagedService.WeightWithoutWaiting(60, "--duration=1m")
			time.Sleep(5 * time.Second)
			stagedComputedWeightAfter10Seconds, serviceComputedWeightAfter10Seconds := stagedService.GetComputedWeight(), service.GetComputedWeight()
			// Wide range given here because the weight change amount is about 500 to 600, so it can be as high as 1601 and as low as 500 in this scenario.
			// Usually the weight will be 1033
			assert.Less(t, 499, stagedComputedWeightAfter10Seconds)
			assert.Greater(t, 1602, stagedComputedWeightAfter10Seconds)
			assert.Equal(t, 10000, serviceComputedWeightAfter10Seconds)
		})
	}, spec.Parallel())
}
