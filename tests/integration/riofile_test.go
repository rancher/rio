package integration

import (
	"fmt"
	"testing"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/tests/testutil"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
)

func riofileTests(t *testing.T, when spec.G, it spec.S) {

	when("A riofile is imported", func() {

		var riofile testutil.TestRiofile

		it.Before(func() {
			riofile.Up(t, "riofile-export.yaml", "")
		})

		it.After(func() {
			riofile.Remove()
		})

		it("should correctly build the system", func() {
			// Export check
			// todo: fix during https://github.com/rancher/rio/issues/641
			//assert.Equal(t, strings.Trim(riofile.Readfile(), "\n"), strings.Trim(riofile.ExportStack(), "\n"), "should have stack export be same as original file")
			// external services
			externalFoo := testutil.GetExternalService(t, "es-foo")
			assert.Equal(t, "www.example.com", externalFoo.GetFQDN(), "should have external service with fqdn")
			assert.Equal(t, externalFoo.GetKubeFQDN(), externalFoo.GetFQDN(), "should have external service with fqdn from k8s")
			externalBar := testutil.GetExternalService(t, "es-bar")
			assert.Equal(t, "1.1.1.1", externalBar.GetFirstIPAddress(), "should have external service with ip")
			assert.Equal(t, externalBar.GetFirstIPAddress(), externalBar.GetKubeFirstIPAddress(), "should have external service with ip from k8s")
			// services and their endpoints
			serviceV0 := testutil.GetService(t, "export-test-image", "export-test-image", "v0")
			serviceV3 := testutil.GetService(t, "export-test-image", "export-test-image", "v3")
			assert.Equal(t, serviceV0.GetEndpointResponse(), "Hello World", "should have service v0 with endpoint")
			assert.Equal(t, serviceV3.GetEndpointResponse(), "Hello World v3", "should have service v3 with endpoint")
			runningPods := serviceV0.GetRunningPods()
			assert.Len(t, runningPods, 2, "There should be 2 pods associated with the service's app since there are two revisions at scale of 1")
			for _, pod := range runningPods {
				assert.Contains(t, pod, serviceV0.App)
				assert.Contains(t, pod, "2/2")
			}
			// routers and their endpoints
			routerFooV0 := testutil.GetRoute(t, "route-foo", "/v0")
			routerFooV3 := testutil.GetRoute(t, "route-foo", "/v3")
			assert.Equal(t, "Hello World", routerFooV0.GetEndpointResponse(), "should have working route paths for v1")
			assert.Equal(t, "Hello World v3", routerFooV3.GetEndpointResponse(), "should have working route paths for v3")
			assert.Equal(t, "Hello World", routerFooV0.GetKubeEndpointResponse())
			assert.Equal(t, "Hello World v3", routerFooV3.GetKubeEndpointResponse())
		})

	}, spec.Parallel())

	when("A riofile with arbitrary k8s manifests", func() {
		var riofile testutil.TestRiofile
		var frontendRoute testutil.TestRoute
		const riofileName = "test-riofile-prune"
		it.Before(func() {
			riofile.Up(t, "riofile-with-K8s-manifests-before.yaml", riofileName)
		})

		it.After(func() {
			riofile.Remove()
		})

		it("should bringup the k8s sample app", func() {
			// Export check
			// todo: fix during https://github.com/rancher/rio/issues/641
			//assert.Equal(t, strings.Trim(riofile.Readfile(), "\n"), strings.Trim(riofile.ExportStack(), "\n"), "should have stack export be same as original file")

			// check frontend service came up
			frontendSvc := testutil.TestService{
				Name:    "frontend",
				App:     "frontend",
				Service: riov1.Service{},
				Version: "",
				T:       t,
			}
			frontendRoute.Add(t, "", "", "to", frontendSvc)
			frontendRoute.GetEndpointResponse()
			//nginx service
			nginxsvc := testutil.GetService(t, "nginx", "nginx", "v0")
			nginxsvc.GetEndpointResponse()

			//assert k8s services came up outside of riofile
			riofile.Up(t, "riofile-with-K8s-manifests-after.yaml", riofileName)
			defer riofile.Remove()

			//check that services/pods/deployments are removed
			_, err := testutil.WaitForNoResponse(fmt.Sprintf("%s%s", frontendRoute.Router.Status.Endpoints[0], frontendRoute.Path))
			assert.Nil(t, err, "the endpoint should go down")
			_, err = testutil.KubectlCmd([]string{"get", "-n", testutil.TestingNamespace, "svc", "frontend"})
			assert.Error(t, err, "kubectl should return an error since the service is being removed")
		}, spec.Parallel())
	})
}
