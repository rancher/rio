package testutil

import (
	"errors"
	"os/exec"
	"strings"
)

// RioCmd func calls the rio command with your arguments
// name=run and args=["-n", "test"] would run: "rio run -n test"
func RioCmd(name string, args []string) (string, error) {
	outBuffer := &strings.Builder{}
	errBuffer := &strings.Builder{}
	args = append([]string{name}, args...) // named command is always first arg
	cmd := exec.Command("rio", args...)
	cmd.Stdout = outBuffer
	cmd.Stderr = errBuffer
	err := cmd.Run()
	if err != nil {
		return outBuffer.String(), errors.New(errBuffer.String())
	}
	return outBuffer.String(), nil
}
