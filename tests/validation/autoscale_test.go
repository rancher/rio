package validation

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

	when("run a service with minscale of 0", func() {
		it.Before(func() {
			service.Create(t, "--scale", "0-4", "strongmonkey1992/autoscale:testing")
		})

		it("should autoscale down to 0", func() {
			// Precondition
			assert.True(t, service.IsReady())
			assert.Equal(t, "Hello World", service.GetAppEndpointResponse())

			// When no requests happen for a while, it should scale to 0
			service.WaitForScaleDown()
			runningPods, availableReplicas := service.GetPodsAndReplicas()
			assert.Equal(t, 0, availableReplicas, "should have 0 available replicas.")
			assert.Len(t, runningPods, 0)

			// Send a request and validate it still executes properly and makes one replica and pod available
			assert.Equal(t, "Hi there, I am rio:production6", service.GetAppEndpointResponse())
			assert.True(t, service.IsReady())
			runningPods, availableReplicas = service.GetPodsAndReplicas()
			assert.Equal(t, 1, availableReplicas, "should have 1 available replica after the endpoint is called once.")
			assert.Len(t, runningPods, 1)
			for _, pod := range runningPods {
				assert.Contains(t, pod, service.Service.Spec.App)
				assert.Contains(t, pod, "2/2")
			}
		})
	}, spec.Parallel())

	when("run a service with a scale range", func() {
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

	when("run a service from github with a fixed scale", func() {

		it.Before(func() {
			service.Create(t, "--scale", "3", "-p", "8080", "https://github.com/rancher/rio-demo")
		})

		it("should have the specified scale, no autoscaling, and be able to have its scale manually adjusted", func() {
			assert.Equal(t, 3, service.GetAvailableReplicas(), "should have three available replicas")
			assert.Equal(t, "Hi there, I'm running in Rio", service.GetAppEndpointResponse())

			service.GenerateLoad("30s", 300)
			runningPods, availableReplicas := service.GetPodsAndReplicas()
			assert.Equal(t, 3, availableReplicas, "should have three available replicas")
			assert.Len(t, runningPods, 3)
			for _, pod := range runningPods {
				assert.Contains(t, pod, service.Service.Name)
			}

			service.Scale(1)
			runningPods, availableReplicas = service.GetPodsAndReplicas()
			assert.Equal(t, 1, availableReplicas, "should have one available replica")
			assert.Equal(t, 1, service.GetScale())
			assert.Len(t, runningPods, 1)
			for _, pod := range runningPods {
				assert.Contains(t, pod, service.Service.Name)
			}
		})
	}, spec.Parallel())
}
