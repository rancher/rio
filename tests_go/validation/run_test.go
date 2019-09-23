// +build validation

package validation

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rancher/rio/tests_go/testutil"
)

var serviceName string

var _ = Describe("Run", func() {
	BeforeEach(func() {
		serviceName = testutil.SetupService()
	})

	Describe("service", func() {
		Context("From scrach", func() {
			It("Should work", func() {
				s, err := testutil.InspectService(serviceName)
				if err != nil {
					Fail(err.Error())
				}
				generatedName := fmt.Sprintf("%s/%s", s.ObjectMeta.Namespace, s.ObjectMeta.Name)
				Expect(generatedName).To(Equal(serviceName))
			})
		})
	})

	AfterEach(func() {
		testutil.CleanupService(serviceName)
	})
})
