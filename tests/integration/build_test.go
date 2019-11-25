package integration

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"

	"github.com/rancher/rio/tests/testutil"
)

func buildTests(t *testing.T, when spec.G, it spec.S) {

	var service testutil.TestService
	var stagedService testutil.TestService

	when("an image is built locally in rio", func() {
		it.Before(func() {
			service.BuildAndCreate(t, "localtest", "v1", "--no-cache", "--build-arg", "VERSION=v1")
			stagedService = service.BuildAndStage(t, "localtest", "v2", "--no-cache", "--build-arg", "VERSION=v2")
		})

		it.After(func() {
			service.Remove()
			stagedService.Remove()
		})

		it("should be able to run a full workload flow with versioning", func() {
			assert.Equal(t, 1, service.GetAvailableReplicas(), "should have one available replica")
			assert.Equal(t, "hello world: v1", service.GetEndpointResponse())
			assert.Equal(t, 1, stagedService.GetAvailableReplicas(), "should have one available replica")
			assert.Equal(t, "hello world: v2", stagedService.GetEndpointResponse())
			assert.Equal(t, "hello world: v1", stagedService.GetAppEndpointResponse())
			stagedService.Promote()
			assert.Equal(t, "hello world: v2", stagedService.GetAppEndpointResponse())
		})
	}, spec.Parallel())
}
