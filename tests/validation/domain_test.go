package validation

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"

	"github.com/rancher/rio/tests/testutil"
)

func domainTests(t *testing.T, when spec.G, it spec.S) {

	var service testutil.TestService
	var domain testutil.TestDomain
	var response *route53.ChangeResourceRecordSetsOutput

	it.Before(func() {
		service.Create(t, "ibuildthecloud/demo:v1")
	})

	it.After(func() {
		service.Remove()
	})

	when("a cname is registered with a service", func() {
		var domainName string

		it.Before(func() {
			domainName, _ = service.GetKubeFirstClusterDomain()
			response = testutil.CreateRoute53DNS(t, "CNAME", domainName)
		})

		it.After(func() {
			domain.UnRegister()
			if os.Getenv("RIO_CNAME") != "" {
				testutil.DeleteRoute53DNS(domainName)
			}
		})

		it("should allow retrieval of the service content using it", func() {
			assert.Equal(t, "DNS for custom record in Rio Cluster Domains", *response.ChangeInfo.Comment)
			domain.RegisterDomain(t, testutil.GetCNAMEInfo(), service.Name)
			assert.Equal(t, testutil.GetCNAMEInfo(), domain.GetDomain())
			assert.Equal(t, testutil.GetCNAMEInfo(), domain.GetKubeDomain())
			urlHTTPResponse, _ := testutil.WaitForURLResponse("http://" + testutil.GetCNAMEInfo())
			assert.Equal(t, "Hello World", urlHTTPResponse)
			urlHTTPSResponse, _ := testutil.WaitForURLResponse("https://" + testutil.GetCNAMEInfo())
			assert.Equal(t, "Hello World", urlHTTPSResponse)
		})
	}, spec.Parallel())

	when("a clusterdomain is added", func() {
		var ip string
		var otherService testutil.TestService

		it.Before(func() {
			_, ip = service.GetKubeFirstClusterDomain()
			response = testutil.CreateRoute53DNS(t, "A", ip)
			testutil.CreateSelfSignedCert(t)
			_, err := testutil.KubectlCmd([]string{"-n", "rio-system", "create", "secret", "tls",
				testutil.GetAInfo() + "-tls", "--cert=testcert.pem", "--key=testkey.pem"})
			if err != nil {
				assert.Fail(t, "Failed to create tls secret for custom cluster domain: "+err.Error())
			}
			domain.ApplyClusterDomain(ip)
			otherService.Create(t, "ibuildthecloud/demo:v3")
		})

		it.After(func() {
			otherService.Remove()
			if os.Getenv("RIO_A_RECORD") != "" {
				testutil.DeleteRoute53DNS(ip)
			}
			_, _ = testutil.KubectlCmd([]string{"-n", "rio-system", "delete", "secret", testutil.GetAInfo() + "-tls"})
			_, _ = testutil.KubectlCmd([]string{"delete", "clusterdomain", testutil.GetAInfo()})
			testutil.DeleteSelfSignedCert()
		})

		it("should add the domain to all rio services", func() {
			assert.Equal(t, "DNS for custom record in Rio Cluster Domains", *response.ChangeInfo.Comment)
			expectedServiceEndpoint := service.Name + "-v0-" + testutil.TestingNamespace + "." + testutil.GetAInfo()
			expectedOtherEndpoint := otherService.Name + "-v0-" + testutil.TestingNamespace + "." + testutil.GetAInfo()
			assert.Contains(t, service.GetEndpointURLs(), "https://"+expectedServiceEndpoint)
			assert.Contains(t, service.GetEndpointURLs(), "http://"+expectedServiceEndpoint)
			assert.Contains(t, otherService.GetEndpointURLs(), "https://"+expectedOtherEndpoint)
			assert.Contains(t, otherService.GetEndpointURLs(), "http://"+expectedOtherEndpoint)
			serviceHTTPResponse, _ := testutil.WaitForURLResponse("http://" + expectedServiceEndpoint)
			assert.Equal(t, "Hello World", serviceHTTPResponse)
			serviceHTTPSResponse, _ := testutil.WaitForURLResponse("https://" + expectedServiceEndpoint)
			assert.Equal(t, "Hello World", serviceHTTPSResponse)
			otherHTTPResponse, _ := testutil.WaitForURLResponse("http://" + expectedOtherEndpoint)
			assert.Equal(t, "Hello World v3", otherHTTPResponse)
			otherHTTPSResponse, _ := testutil.WaitForURLResponse("https://" + expectedOtherEndpoint)
			assert.Equal(t, "Hello World v3", otherHTTPSResponse)
		})
	}, spec.Parallel())
}
