package integration

import (
	"testing"

	"github.com/rancher/rio/tests/testutil"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
)

func configTests(t *testing.T, when spec.G, it spec.S) {

	var config testutil.TestConfig

	it.After(func() {
		config.Remove()
	})

	when("a config is created with data", func() {
		it("should contain that data", func() {
			testText := []string{"a=b", "foo=bar"}
			config.Create(t, testText)
			assert.Equal(t, testText, config.GetContent())
			assert.Equal(t, testText, config.GetKubeContent())
		})
	}, spec.Parallel())
}
