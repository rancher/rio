package integration

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"

	"github.com/rancher/rio/tests/testutil"
)

func runTests(t *testing.T, when spec.G, it spec.S) {

	var service testutil.TestService

	it.Before(func() {
		service.Create(t, "nginx")
	})

	it.After(func() {
		service.Remove()
	})

	when("rio run is called", func() {
		it("should create a service with default specifications", func() {
			runningPods := service.GetRunningPods()
			assert.Equal(t, 1, service.GetAvailableReplicas(), "should have one available replica")
			assert.Equal(t, 100, service.GetCurrentWeight())
			assert.Equal(t, "nginx", service.GetImage())
			assert.Contains(t, runningPods, service.App.Name)
			assert.Contains(t, runningPods, "2/2")
		})
	})
}
