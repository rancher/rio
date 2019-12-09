package integration

import (
	"fmt"
	"reflect"
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
			export, err := riofile.ExportStack()
			assert.Nil(t, err)
			orig, err := riofile.Readfile()
			assert.Nil(t, err)
			assert.True(t, reflect.DeepEqual(orig["services"], export["services"]), "export should have same services as file")
			assert.True(t, reflect.DeepEqual(orig["externalservices"], export["externalservices"]), "export should have same externalservices as file")
			assert.True(t, reflect.DeepEqual(orig["routers"], export["routers"]), "export should have same routers as file")
			assert.Contains(t, export, "kubernetes")

			// external service FQDN
			externalFoo := testutil.GetExternalService(t, "es-foo")
			externalFQDN := externalFoo.GetFQDN()
			assert.Equal(t, "www.example.com", externalFQDN, "should have external service with fqdn")
			assert.Equal(t, externalFQDN, externalFoo.GetKubeFQDN(), "should have external service with fqdn from k8s")

			// external service IP
			externalBar := testutil.GetExternalService(t, "es-bar")
			externalIP := externalBar.GetFirstIPAddress()
			assert.Equal(t, "1.1.1.1", externalIP, "should have external service with ip")
			assert.Equal(t, externalIP, externalBar.GetKubeFirstIPAddress(), "should have external service with ip from k8s")

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
			frontendRoute.Remove()
		})

		it("should bring up the k8s sample app", func() {
			// Export check, ensure manifest services end up as rio services
			export, err := riofile.ExportStack()
			assert.Nil(t, err)
			exportSvcs := export["services"].(map[string]interface{})
			assert.Contains(t, exportSvcs, "nginx")

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
			_, err = testutil.WaitForNoResponse(fmt.Sprintf("%s%s", frontendRoute.Router.Status.Endpoints[0], frontendRoute.Path))
			assert.Nil(t, err, "the endpoint should go down")
			_, err = testutil.KubectlCmd([]string{"get", "-n", testutil.TestingNamespace, "svc", "frontend"})
			assert.Error(t, err, "kubectl should return an error since the service is being removed")
		}, spec.Parallel())
	})
}
