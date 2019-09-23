package validation

import (
	"testing"

	"github.com/sclevine/spec"
)

func logTests(t *testing.T, when spec.G, it spec.S) {

	when("rio logs is called", func() {
		it("should respond with logs", func() {
			if 1 == 2 {
				t.Error("bad")
			}
		})
	})
}
