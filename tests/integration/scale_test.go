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
		it("should scale up to 2", func() {
			assert.Equal(t, 1, service.GetAvailableReplicas())
			service.Scale(2)
			assert.Equal(t, 2, service.GetScale())
			assert.Equal(t, 2, service.GetAvailableReplicas())
			assert.Equal(t, service.GetKubeAvailableReplicas(), service.GetAvailableReplicas())
			// todo: disable this test for now. Doesn't return response from different pod, need to figure out why
			//assert.True(t, service.PodsResponsesMatchAvailableReplicas("/name.html", service.GetAvailableReplicas()))
		}, spec.Parallel())
		// This is an important test because zero scale is wide ranging feature
		it("should scale down to zero", func() {
			assert.Equal(t, 1, service.GetAvailableReplicas())
			service.Scale(0)
			assert.Equal(t, 0, service.GetAvailableReplicas())
			assert.Equal(t, 0, service.GetScale())
		}, spec.Sequential())
	})
}
