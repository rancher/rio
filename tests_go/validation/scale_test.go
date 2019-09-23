// +build validation

package validation

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rancher/rio/tests_go/testutil"
)

var _ = Describe("Scale", func() {
	var serviceName string

	BeforeEach(func() {
		serviceName = testutil.SetupService()
	})

	Describe("Service", func() {
		Context("With an already running service", func() {
			It("Should scale up correctly", func() {
				Expect(getScale(serviceName)).To(Equal(1))
				setScale(serviceName, 2)
				waitForScale(serviceName, 2)
				Expect(getScale(serviceName)).To(Equal(2))
			})
			It("Should scale down correctly", func() {
				Expect(getScale(serviceName)).To(Equal(1))
				setScale(serviceName, 0)
				waitForScale(serviceName, 0)
				Expect(getScale(serviceName)).To(Equal(0))
			})
		})
	})

	AfterEach(func() {
		testutil.CleanupService(serviceName)
	})
})

func setScale(serviceName string, scaleTo int) (string, error) {
	args := []string{
		fmt.Sprintf("%s=%d", serviceName, scaleTo),
	}
	out, err := testutil.RioCmd("scale", args)
	return out, err
}

func getScale(serviceName string) (int, error) {
	out, err := testutil.InspectService(serviceName)
	if err != nil {
		return -1, err
	}
	if out.Status.ScaleStatus != nil {
		return out.Status.ScaleStatus.Ready, nil
	}
	return -1, errors.New("unable to get scale")
}

func waitForScale(serviceName string, wantScale int) {
	f := func() bool {
		count, err := getScale(serviceName)
		if err == nil && count == wantScale {
			return true
		}
		return false
	}
	o := testutil.WaitFor(f, 30)
	if o == false {
		Fail("Timed out scaling service")
	}
}
