package funcs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitPreserveQuotes(t *testing.T) {
	assert.Equal(t, []string{
		"--a",
	}, SplitPreserveQuotes("--a"))
	assert.Equal(t, []string{
		"--a",
		"--b",
	}, SplitPreserveQuotes("--a --b"))
	assert.Equal(t, []string{
		"--a",
		"--b='c d'",
		"--e='f'",
	}, SplitPreserveQuotes("--a --b='c d' --e='f'"))
	assert.Equal(t, []string{
		"--a",
		"--b=\"c d\"",
		"--e=\"f\"",
	}, SplitPreserveQuotes("--a --b=\"c d\" --e=\"f\""))
}
