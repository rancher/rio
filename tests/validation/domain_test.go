package validation

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"

	"github.com/rancher/rio/tests/testutil"
)

func domainTests(t *testing.T, when spec.G, it spec.S) {

	var service testutil.TestService
	var domain testutil.TestDomain

	it.After(func() {
		service.Remove()
	})

	when("get a cluster domain", func() {
		it.Before(func() {
			service.Create(t, "strongmonkey1992/autoscale:testing")
		})

		it.After(func() {
			domain.UnRegister()
			testutil.DeleteCNAME(service.GetKubeFirstClusterDomain())
		})

		it("should create a new DNS CNAME for the cluster domain and retrieve the service content using it", func() {
			response := testutil.CreateCNAME(service.GetKubeFirstClusterDomain())
			assert.Equal(t, *response.ChangeInfo.Comment, "Add CNAME to Rio Cluster Domain")
			domain.RegisterDomain(t, testutil.GetCNAMEInfo(), service.Name)
			assert.Equal(t, testutil.GetCNAMEInfo(), domain.GetDomain())
			assert.Equal(t, testutil.GetCNAMEInfo(), domain.GetKubeDomain())
			urlResponse, _ := testutil.WaitForURLResponse("https://" + testutil.GetCNAMEInfo())
			assert.Equal(t, "Hi there, I am rio:production6", urlResponse)
		})
	})
}
