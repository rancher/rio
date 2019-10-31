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

	it.Before(func() {
		service.Create(t, "--label", "x=y", "--annotations", "a=b", "ibuildthecloud/demo:v1")
	})

	it.After(func() {
		service.Remove()
		stagedService.Remove()
	})

	when("a running service has a version staged", func() {
		it("should have proper fields assigned", func() {
			stageVersion := "v3"
			stagedService = service.Stage("ibuildthecloud/demo:v3", stageVersion)
			stageName := fmt.Sprintf("%s-%s", service.App, stageVersion)
			assert.Equal(t, stageName, stagedService.Name, "should have correct name")
			assert.Equal(t, 0, stagedService.GetSpecWeight(), "should have initial weight set to 0")
			assert.Equal(t, stageVersion, stagedService.Service.Spec.Version, "should have supplied version")
			assert.Equal(t, service.Service.Annotations, stagedService.Service.Annotations, "should copy annotations")
			assert.Equal(t, service.Service.Labels, stagedService.Service.Labels, "should copy labels")
			assert.Equal(t, map[string]string{"x": "y"}, stagedService.Service.Labels, "should have correct labels")
		})
	}, spec.Parallel())
}
