package integration

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"

	"github.com/rancher/rio/tests/testutil"
)

func scaleTests(t *testing.T, when spec.G, it spec.S) {

	var service testutil.TestService

	it.Before(func() {
		service.Create(t, "nginx")
	})

	it.After(func() {
		service.Remove()
	})

	when("scale is called on a service", func() {
		it("should scale up", func() {
			assert.Equal(t, 1, service.GetAvailableReplicas())
			service.Scale(2)
			assert.Equal(t, 2, service.GetAvailableReplicas())
		})
		// This is an important test because zero scale is wide ranging feature
		it("should scale to zero", func() {
			assert.Equal(t, 1, service.GetAvailableReplicas())
			service.Scale(0)
			assert.Equal(t, 0, service.GetAvailableReplicas())
			assert.Equal(t, 0, service.GetScale())
		})
		it.Pend("should scale to x,y,z....", func() {
		})
	}, spec.Parallel())
}
