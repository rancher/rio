package integration

import (
	"testing"

	"github.com/rancher/rio/tests/testutil"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
)

func exportTests(t *testing.T, when spec.G, it spec.S) {

	serviceImage := "nginx"
	var service testutil.TestService

	it.Before(func() {
		service.Create(t, serviceImage)
	})

	it.After(func() {
		service.Remove()
	})

	when("A service is running and we call export on it", func() {
		it("should have populated fields in normal format", func() {
			exportedService := service.Export()
			assert.Equal(t, serviceImage, exportedService.GetImage(), "should have correct image")
			assert.Equal(t, 1, exportedService.GetScale(), "should have scale of 1")
		})
		it("should have populated fields in raw format", func() {
			exportedService := service.ExportRaw()
			assert.Equal(t, serviceImage, exportedService.GetImage(), "should have correct image")
			assert.Equal(t, 1, exportedService.GetScale(), "should have scale of 1")
		})
	}, spec.Parallel())
}
