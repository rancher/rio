package logs_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logs", func() {
	Describe("", func() {
		Context("With more than 100 lines", func() {
			It("should equal 1", func() {
				Expect(1).To(Equal(1))
			})
		})
	})
})
