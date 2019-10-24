package validation

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"

	"github.com/rancher/rio/tests/testutil"
)

func autoscaleTests(t *testing.T, when spec.G, it spec.S) {

	var service testutil.TestService

	when("run a service with minscale of 0", func() {
		it.Before(func() {
			service.Create(t, "--scale", "0-4", "ibuildthecloud/demo:v1")
		})
		it.After(func() {
			service.Remove()
		})

		it("should autoscale down to 0", func() {
			// Precondition
			assert.Equal(t, 1, service.GetAvailableReplicas(), "should have one initial replica. Failed Precondition")
			assert.Equal(t, "Hello World", service.GetAppEndpointResponse())

			// When no requests happen for a while, it should scale to 0
			service.WaitForScaleDown()
			assert.Equal(t, 0, service.GetAvailableReplicas(), "should have 0 available replicas.")
			runningPods := service.GetRunningPods()
			assert.Len(t, runningPods, 0)

			// Send a request and validate it still executes properly and makes one replica and pod available
			assert.Equal(t, "Hello World", service.GetAppEndpointResponse())
			assert.Equal(t, 1, service.GetAvailableReplicas(), "should have 1 available replica after the endpoint is called once.")
			runningPods = service.GetRunningPods()
			assert.Len(t, runningPods, 1)
			for _, pod := range runningPods {
				assert.Contains(t, pod, service.Service.Spec.App)
				assert.Contains(t, pod, "2/2")
			}
		})
	}, spec.Parallel())
}
