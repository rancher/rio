package scale_test

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	testUtils "github.com/rancher/rio/tests"
)

func TestScale(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Scale Suite")
}

func setScale(workloadName string, scaleTo int) (string, error) {
	args := []string{
		fmt.Sprintf("%s=%d", workloadName, scaleTo),
	}
	out, err := testUtils.RioCmd("scale", args)
	return out, err
}

func getScale(workloadName string) (int, error) {
	out, err := testUtils.InspectService(workloadName)
	if err != nil {
		return -1, err
	}
	if out.Status.ScaleStatus != nil {
		return out.Status.ScaleStatus.Ready, nil
	}
	return -1, errors.New("unable to get scale")
}

func waitForScale(workloadName string, wantScale int) {
	f := func() bool {
		count, err := getScale(workloadName)
		if err == nil && count == wantScale {
			return true
		}
		return false
	}
	o := testUtils.WaitFor(f, 30)
	if o == false {
		Fail("Timed out scaling workload")
	}
}
