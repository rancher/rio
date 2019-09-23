// +build validation

package validation

import (
	"fmt"
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"

	"github.com/rancher/rio/tests_go/testutil"
)

func runTests(t *testing.T, when spec.G, it spec.S) {

	var serviceName string

	it.Before(func() {
		serviceName = testutil.SetupService()
	})

	it.After(func() {
		testutil.CleanupService(serviceName)
	})

	when("rio run", func() {
		it("should do one thing up", func() {
			s, err := testutil.InspectService(serviceName)
			if err != nil {
				t.Error(err.Error())
			}
			generatedName := fmt.Sprintf("%s/%s", s.ObjectMeta.Namespace, s.ObjectMeta.Name)
			assert.Equal(t, generatedName, serviceName)
		})
	})
}
