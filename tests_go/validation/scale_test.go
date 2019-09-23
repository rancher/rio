package validation

import (
	"errors"
	"fmt"
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"

	"github.com/rancher/rio/tests_go/testutil"
)

func scaleTests(t *testing.T, when spec.G, it spec.S) {

	var serviceName string

	it.Before(func() {
		serviceName = testutil.SetupService()
	})

	it.After(func() {
		testutil.CleanupService(serviceName)
	})

	when("A service is already running", func() {
		it("should scale up correctly", func() {
			currScale, _ := getScale(serviceName)
			assert.Equal(t, currScale, 1)
			setScale(serviceName, 2)
			waitForScale(t, serviceName, 2)
			currScale, err := getScale(serviceName)
			if err != nil {
				t.Logf(err.Error())
				t.Fail()
			}
			assert.Equal(t, currScale, 2)
		})
		it("should scale down correctly", func() {
			currScale, _ := getScale(serviceName)
			assert.Equal(t, currScale, 1)
			setScale(serviceName, 0)
			waitForScale(t, serviceName, 0)
			currScale, err := getScale(serviceName)
			if err != nil {
				t.Logf(err.Error())
				t.Fail()
			}
			assert.Equal(t, currScale, 0)
		})
	}, spec.Parallel())
}

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

func waitForScale(t *testing.T, serviceName string, wantScale int) {
	f := func() bool {
		count, err := getScale(serviceName)
		if err == nil && count == wantScale {
			return true
		}
		return false
	}
	o := testutil.WaitFor(f, 30)
	if o == false {
		t.Logf("Timed out scaling service")
		t.Fail()
	}
}
