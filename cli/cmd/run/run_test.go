package run_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	testUtils "github.com/rancher/rio/tests"
)

var workloadName string

var _ = Describe("Run", func() {
	BeforeEach(func() {
		workloadName = testUtils.SetupWorkload()
	})

	Describe("Workload", func() {
		Context("From scrach", func() {
			It("Should work", func() {
				s, err := testUtils.InspectService(workloadName)
				if err != nil {
					Fail(err.Error())
				}
				Expect(s.ObjectMeta.Name).To(Equal(workloadName))
			})
		})
	})

	AfterEach(func() {
		testUtils.CleanupWorkload(workloadName)
	})
})
