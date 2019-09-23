// +build validation

package validation

import (
	"testing"

	"github.com/sclevine/spec"
)

func logTests(t *testing.T, when spec.G, it spec.S) {

	it("should have a test", func() {
		if 1 == 2 {
			t.Error("bad")
		}
	})
}
