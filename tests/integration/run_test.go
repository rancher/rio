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

	when("run a service with liveness check", func() {
		it("should come become available", func() {
			service.Create(t, "-p", "80", "--health-url", "http://:80", "--health-initial-delay", "1s", "--health-interval", "1s", "--health-failure-threshold", "1", "--health-timeout", "1s", "ibuildthecloud/demo:v1")
			assert.Equal(t, 1, service.GetAvailableReplicas(), "should have one available replica")
			assert.Equal(t, "Hello World", service.GetAppEndpointResponse())
		})
	}, spec.Parallel())

	when("run a service from a public github repository", func() {
		it("should come become available", func() {
			service.Create(t, "-p", "8080", "https://github.com/rancher/rio-demo")
			assert.Equal(t, 1, service.GetAvailableReplicas(), "should have one available replica")
			assert.Equal(t, "Hi there, I'm running in Rio", service.GetAppEndpointResponse())
		})
	}, spec.Parallel())

	when("run a service that can expose multiple ports", func() {
		it("should expose the port that is listed last", func() {
			service.Create(t, "-p", "8002:8002", "-p", "8001:8001", "maxross/multiport:testing")
			assert.Equal(t, 1, service.GetAvailableReplicas(), "should have one available replica")
			assert.Equal(t, "Listening on 8001", service.GetAppEndpointResponse())
		})

		it("should not expose any internal ports", func() {
			service.Create(t, "-p", "8002:8002", "-p", "8001:8001,internal=true", "maxross/multiport:testing")
			assert.Equal(t, 1, service.GetAvailableReplicas(), "should have one available replica")
			assert.Equal(t, "Listening on 8002", service.GetAppEndpointResponse())
		})
	}, spec.Parallel())
}
