package integration

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"

	"github.com/rancher/rio/tests/testutil"
)

func autoscaleTests(t *testing.T, when spec.G, it spec.S) {

	var service testutil.TestService

	it.After(func() {
		service.Remove()
	})

	when("rio run is called with a scale range", func() {

		it.Before(func() {
			service.Create(t, "--scale", "1-10", "strongmonkey1992/autoscale:testing")
		})

		it("should create a service that can scale up", func() {
			assert.Equal(t, 1, service.GetAvailableReplicas(), "should have one available replica")
			assert.Equal(t, 100, service.GetCurrentWeight())
			assert.Equal(t, "strongmonkey1992/autoscale:testing", service.GetImage())
			assert.Contains(t, service.Exec("env"), "KUBERNETES_SERVICE_PORT")

			service.GenerateLoad("10s", 300)
			runningPods, availableReplicas := service.GetPodsAndReplicas()
			assert.Greater(t, len(runningPods), 1)
			assert.Greater(t, availableReplicas, 1, "should have more than 1 available replica")
		})
	}, spec.Parallel())
}
