package integration

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"

	"github.com/rancher/rio/tests/testutil"
)

func logTests(t *testing.T, when spec.G, it spec.S) {

	var service testutil.TestService

	when("run a service whose container prints logs", func() {
		it.Before(func() {
			service.Create(t, "izaac/attachtest:v1")
		})
		it.After(func() {
			service.Remove()
		})

		it("should be able to see the logs for that service", func() {
			logResults := service.Logs()
			assert.Greater(t, len(logResults), 1)
			for i, str := range logResults {
				if i == 0 {
					assert.Contains(t, str, service.Name)
					assert.Contains(t, str, "+ "+service.Service.Namespace)
				} else {
					assert.Contains(t, str, "UTC")
					assert.Contains(t, str, service.Name)
				}
			}
		})
	}, spec.Parallel())
}
