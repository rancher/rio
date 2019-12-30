package integration

import (
	"fmt"
	"testing"

	"github.com/rancher/rio/tests/testutil"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
)

func stageTests(t *testing.T, when spec.G, it spec.S) {

	var service testutil.TestService
	var stagedService testutil.TestService

	when("a running service has a version staged", func() {
		var stageVersion string

		it.Before(func() {
			service.Create(t, "--weight", "100", "--label", "x=y", "--annotations", "a=b", "ibuildthecloud/demo:v1")
			stageVersion = "v3"
			stagedService = service.Stage("ibuildthecloud/demo:v3", stageVersion)
		})

		it.After(func() {
			service.Remove()
			stagedService.Remove()
		})

		it("should have proper fields assigned", func() {
			stageName := fmt.Sprintf("%s@%s", service.App, stageVersion)
			assert.Equal(t, stageName, stagedService.Name, "should have correct name")
			assert.Equal(t, 0, stagedService.GetSpecWeight(), "should have initial weight set to 0")
			assert.Equal(t, stageVersion, stagedService.Service.Spec.Version, "should have supplied version")
			assert.Equal(t, service.Service.Annotations, stagedService.Service.Annotations, "should copy annotations")
			assert.Equal(t, service.Service.Labels, stagedService.Service.Labels, "should copy labels")
			assert.Equal(t, map[string]string{"x": "y"}, stagedService.Service.Labels, "should have correct labels")
		})
		it("should have individual endpoints and an app endpoint pointing to the first version", func() {
			assert.Equal(t, "Hello World", service.GetEndpointResponse())
			assert.Equal(t, "Hello World v3", stagedService.GetEndpointResponse())
			assert.Equal(t, "Hello World", service.GetAppEndpointResponse(), "Response should only be from original service")
			assert.Equal(t, testutil.GetHostname(service.GetEndpointURLs()[0]), testutil.GetHostname(service.GetKubeEndpointURLs()[0]))
			assert.Equal(t, testutil.GetHostname(stagedService.GetEndpointURLs()[0]), testutil.GetHostname(stagedService.GetKubeEndpointURLs()[0]))
			assert.Equal(t, testutil.GetHostname(service.GetAppEndpointURLs()[0]), testutil.GetHostname(service.GetKubeAppEndpointURLs()[0]))
			assert.Equal(t, testutil.GetHostname(service.GetAppEndpointURLs()[0]), testutil.GetHostname(stagedService.GetAppEndpointURLs()[0]))
		})
	}, spec.Parallel())

	when("a running service has a version staged with the run command", func() {

		it.Before(func() {
			service.Create(t, "ibuildthecloud/demo:v1")
		})

		it.After(func() {
			service.Remove()
			stagedService.Remove()
		})

		it("should have proper fields assigned", func() {
			stagedService = service.RunStage("ibuildthecloud/demo:v3", "v3", "80", "50")
			stageName := fmt.Sprintf("%s@%s", service.App, "v3")
			assert.Equal(t, stageName, stagedService.Name, "should have correct name")
			assert.Equal(t, service.App, stagedService.App, "should have same app")
			assert.Equal(t, 10000, stagedService.GetSpecWeight(), "should have weight set") // 10k is to match other 10k at 50%
		})
	}, spec.Parallel())
}
