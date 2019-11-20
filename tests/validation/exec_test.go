package validation

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"

	"github.com/rancher/rio/tests/testutil"
)

func execTests(t *testing.T, when spec.G, it spec.S) {

	var service testutil.TestService
	var otherService testutil.TestService

	when("multiple services are running", func() {
		it.Before(func() {
			service.Create(t, "ibuildthecloud/demo:v1")
			otherService.Create(t, "ibuildthecloud/demo:v3")
		})

		it.After(func() {
			service.Remove()
			otherService.Remove()
		})

		it("should allow accessing services directly from within other service pods in the same namespace", func() {
			assert.Contains(t, service.Exec("/bin/sh", "-c", "apt-get update && apt-get install -y --fix-missing curl && curl "+otherService.Name), "Hello World v3")
		})
	}, spec.Parallel())
}
