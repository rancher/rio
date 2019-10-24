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
		service.Create(t, "sangeetha/mytestcontainer")
	})

	it.After(func() {
		service.Remove()
	})

	when("scale is called on a service", func() {
		it("should scale up to 3", func() {
			assert.Equal(t, 1, service.GetAvailableReplicas())
			service.Scale(3)
			assert.Equal(t, 3, service.GetScale())
			assert.Equal(t, 3, service.GetAvailableReplicas())
			// todo fix below
			//assert.Equal(t, service.GetKubeAvailableReplicas(), service.GetAvailableReplicas())
			//assert.True(t, service.PodsResponsesMatchAvailableReplicas("/name.html", service.GetAvailableReplicas()))
		})
		// This is an important test because zero scale is wide ranging feature
		it("should scale to zero", func() {
			assert.Equal(t, 1, service.GetAvailableReplicas())
			service.Scale(0)
			assert.Equal(t, 0, service.GetAvailableReplicas())
			assert.Equal(t, 0, service.GetScale())
		})
		it("should scale to 11", func() {
			assert.Equal(t, 1, service.GetAvailableReplicas())
			service.Scale(11)
			assert.Equal(t, 11, service.GetAvailableReplicas())
			assert.Equal(t, 11, service.GetScale())
			// todo fix below
			//assert.Equal(t, service.GetKubeAvailableReplicas(), service.GetAvailableReplicas())
			//assert.True(t, service.PodsResponsesMatchAvailableReplicas("/name.html", service.GetAvailableReplicas()))
		})
	}, spec.Parallel())
}
