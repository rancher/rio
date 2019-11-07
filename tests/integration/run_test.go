package integration

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"

	"github.com/rancher/rio/tests/testutil"
)

func runTests(t *testing.T, when spec.G, it spec.S) {

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
			runningPods := service.GetRunningPods()
			for _, pod := range runningPods {
				assert.Contains(t, pod, service.Service.Name)
			}
			// todo: fix flakey test below
			service.GenerateLoad()
			assert.Greater(t, service.GetAvailableReplicas(), 1, "should have more than 1 available replica")
			runningPods = service.GetRunningPods()
			assert.Greater(t, len(runningPods), 1)
		})
	}, spec.Parallel())

	when("run a service from github with a fixed scale", func() {

		it.Before(func() {
			service.Create(t, "--scale", "3", "-p", "8080", "https://github.com/rancher/rio-demo")
		})

		it("should have the specified scale, no autoscaling, and be able to have its scale manually adjusted", func() {
			assert.Equal(t, 3, service.GetAvailableReplicas(), "should have three available replicas")
			assert.Equal(t, "Hi there, I'm running in Rio", service.GetAppEndpointResponse())

			assert.Equal(t, 3, service.GetAvailableReplicas(), "should have three available replicas")
			runningPods := service.GetRunningPods()
			assert.Len(t, runningPods, 3)
			for _, pod := range runningPods {
				assert.Contains(t, pod, service.Service.Name)
			}

			service.Scale(1)
			assert.Equal(t, 1, service.GetAvailableReplicas(), "should have one available replica")
			assert.Equal(t, 1, service.GetScale())
			runningPods = service.GetRunningPods()
			assert.Len(t, runningPods, 1)
			for _, pod := range runningPods {
				assert.Contains(t, pod, service.Service.Name)
			}
		})
	}, spec.Parallel())
}
