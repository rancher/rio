package integration

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"

	"github.com/rancher/rio/tests/testutil"
)

func attachTests(t *testing.T, when spec.G, it spec.S) {

	var service testutil.TestService

	when("run a service whose container prints logs", func() {
		it.Before(func() {
			service.Create(t, "izaac/attachtest:v1")
		})
		it.After(func() {
			service.Remove()
		})

		it("should be able to attach to the service", func() {
			attachResults := service.Attach()
			assert.Greater(t, len(attachResults), 0)
			for _, str := range attachResults {
				assert.Contains(t, str, "UTC")
			}
		})
	}, spec.Parallel())
}
