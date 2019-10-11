package integration

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"

	"github.com/rancher/rio/tests/testutil"
)

func exportTests(t *testing.T, when spec.G, it spec.S) {

	when("A service is running and we call export on it", func() {

		serviceImage := "nginx"
		var service testutil.TestService

		it.Before(func() {
			service.Create(t, serviceImage)
		})

		it.After(func() {
			service.Remove()
		})

		it("should have correct field data", func() {
			exportedService := service.Export()
			assert.Equal(t, serviceImage, exportedService.GetImage(), "should have correct image in standard format")
			assert.Equal(t, 1, exportedService.GetScale(), "should have scale of 1 in standard format")
			rawExportedService := service.ExportRaw()
			assert.Equal(t, serviceImage, rawExportedService.GetImage(), "should have correct image in raw format")
			assert.Equal(t, 1, rawExportedService.GetScale(), "should have scale of 1 in raw format")
		})
	}, spec.Parallel())

}
