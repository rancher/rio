package integration

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"

	"github.com/rancher/rio/tests/testutil"
)

func runTests(t *testing.T, when spec.G, it spec.S) {

	var service testutil.TestService

	it.Before(func() {
		service.Create(t, "nginx")
	})

	it.After(func() {
		service.Remove()
	})

	when("rio run is called", func() {
		it("should create a service with default specifications", func() {
			assert.Equal(t, 1, service.GetAvailableReplicas(), "should have one available replica")
			// TODO: check the following and then remove test_run.py
			// scale of 1
			// weight of 100 in spec
			// image == "nginx"
		})
	})
}
