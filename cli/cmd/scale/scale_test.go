package scale_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testUtils "github.com/rancher/rio/tests"
)

var workloadName string

var _ = Describe("Scale", func() {
	BeforeEach(func() {
		workloadName = testUtils.SetupWorkload()
	})

	Describe("Workload", func() {
		Context("With an already running workload", func() {
			It("Should scale up correctly", func() {
				Expect(getScale(workloadName)).To(Equal(1))
				setScale(workloadName, 2)
				waitForScale(workloadName, 2)
				Expect(getScale(workloadName)).To(Equal(2))
			})
			It("Should scale down correctly", func() {
				Expect(getScale(workloadName)).To(Equal(1))
				setScale(workloadName, 0)
				waitForScale(workloadName, 0)
				Expect(getScale(workloadName)).To(Equal(0))
			})
		})
	})

	AfterEach(func() {
		testUtils.CleanupWorkload(workloadName)
	})
})
