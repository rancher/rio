package integration

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"

	"github.com/rancher/rio/tests/testutil"
)

func riofileTests(t *testing.T, when spec.G, it spec.S) {

	when("A riofile is imported", func() {

		var riofile testutil.TestRiofile

		it.Before(func() {
			riofile.Up(t, "riofile-export.yaml")
		})

		it.After(func() {
			riofile.Remove()
		})

		it("should correctly build the system", func() {
			// Export check
			assert.Equal(t, riofile.Readfile(), riofile.ExportStack(), "should have stack export be same as original file")
			// external services
			externalFoo := testutil.GetExternalService(t, "es-foo")
			assert.Equal(t, "www.example.com", externalFoo.GetFQDN(), "should have external service with fqdn")
			externalBar := testutil.GetExternalService(t, "es-bar")
			assert.Equal(t, "1.1.1.1", externalBar.GetFirstIPAddress(), "should have external service with ip")
			// services and their endpoints
			serviceV0 := testutil.GetService(t, "export-test-image", "v0")
			serviceV3 := testutil.GetService(t, "export-test-image", "v3")
			assert.Equal(t, serviceV0.GetEndpointResponse(), "Hello World", "should have service v0 with endpoint")
			assert.Equal(t, serviceV3.GetEndpointResponse(), "Hello World v3", "should have service v3 with endpoint")
			runningPods := serviceV0.GetRunningPods()
			assert.Len(t, runningPods, 2, "There should be 2 pods associated with the service's app since there are two revisions at scale of 1")
			for _, pod := range runningPods {
				assert.Contains(t, pod, "export-test-image")
				assert.Contains(t, pod, "2/2")
			}
			// routers and their endpoints
			routerBar := testutil.GetRoute(t, "route-bar", "/bv0")
			assert.Equal(t, "/bv0", routerBar.Router.Spec.Routes[0].Match.Path.Exact, "should have correct route set")
			routerFooV0 := testutil.GetRoute(t, "route-foo", "/v0")
			routerFooV3 := testutil.GetRoute(t, "route-foo", "/v3")
			assert.Equal(t, "Hello World", routerFooV0.GetEndpointResponse(), "should have working route paths for v1")
			assert.Equal(t, "Hello World v3", routerFooV3.GetEndpointResponse(), "should have working route paths for v3")
			assert.Equal(t, "Hello World", routerFooV0.GetKubeEndpointResponse())
			assert.Equal(t, "Hello World v3", routerFooV3.GetKubeEndpointResponse())
		})
	}, spec.Parallel())
}
