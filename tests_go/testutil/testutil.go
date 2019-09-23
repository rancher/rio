package testutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os/exec"
	"strings"
	"time"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"

	"github.com/onsi/ginkgo"
)

const testing_namespace = "testing-ns"

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

func SetupService() string {
	rand.Seed(time.Now().UnixNano())
	name := fmt.Sprintf("test-service-%v", rand.Intn(99999))
	name = GetFullServiceName(name)
	args := []string{"-n", name, "nginx"}
	out, err := RioCmd("run", args)
	if err != nil {
		if err.Error() == "" {
			ginkgo.Fail(out)
		}
		ginkgo.Fail(err.Error())
	}
	err = WaitForService(name)
	if err != nil {
		ginkgo.Fail(err.Error())
	}
	return name
}

func CleanupService(name string) {
	_, err := RioCmd("rm", []string{name})
	if err != nil {
		ginkgo.Fail(err.Error())
	}
}

func InspectService(name string) (riov1.Service, error) {
	r := riov1.Service{}
	args := append([]string{"--type", "service", "--format", "json"}, name)
	out, err := RioCmd("inspect", args)
	if err != nil {
		return r, err
	}
	if err := json.Unmarshal([]byte(out), &r); err != nil {
		return r, err
	}
	return r, nil
}

func GetFullServiceName(name string) string {
	return fmt.Sprintf("%s/%s", testing_namespace, name)
}
