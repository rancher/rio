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

	when("rio run is called", func() {

		it.Before(func() {
			service.Create(t)
		})

		it("should create a service with default specifications that can scale up", func() {
			assert.Equal(t, 1, service.GetAvailableReplicas(), "should have one available replica")
			assert.Equal(t, 100, service.GetCurrentWeight())
			assert.Equal(t, "nginx", service.GetImage())
			//assert.Contains(t, service.Logs(), "linkerd-init") // todo: timeout or dont follow logs
			assert.Contains(t, service.Exec("env"), "KUBERNETES_SERVICE_PORT")
			runningPods := service.GetRunningPods()
			for _, pod := range runningPods {
				assert.Contains(t, pod, service.Service.Name)
				assert.Contains(t, pod, "2/2")
			}
			// todo: put scale tests in scale_test.go, both here and below
			service.GenerateLoad() // todo: fix hey cmd in here
			assert.Greater(t, service.GetAvailableReplicas(), 1, "should have more than 1 available replica")
			runningPods = service.GetRunningPods()
			assert.Greater(t, len(runningPods), 1)
		})
	}, spec.Parallel())

	when("run a service with a fixed scale", func() {

		it.Before(func() {
			service.Create(t, "--scale", "3", "https://github.com/rancher/rio-demo")
		})

		it("should have the specified scale, no autoscaling, and be able to have its scale manually adjusted", func() {
			assert.Equal(t, 3, service.GetAvailableReplicas(), "should have three available replicas")
			assert.Equal(t, "Hi there, I'm running in Rio", service.GetAppEndpointResponse())

			service.GenerateLoad()
			assert.Equal(t, 3, service.GetAvailableReplicas(), "should have three available replicas")
			runningPods := service.GetRunningPods()
			assert.Len(t, runningPods, 3)
			for _, pod := range runningPods {
				assert.Contains(t, pod, service.Service.Name)
				assert.Contains(t, pod, "2/2")
			}

			service.Scale(1)
			assert.Equal(t, 1, service.GetAvailableReplicas(), "should have one available replica")
			assert.Equal(t, 1, service.GetScale())
			runningPods = service.GetRunningPods()
			assert.Len(t, runningPods, 1)
			for _, pod := range runningPods {
				assert.Contains(t, pod, service.Service.Name)
				assert.Contains(t, pod, "2/2")
			}
		})
	}, spec.Parallel())
}
