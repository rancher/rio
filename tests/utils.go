package tests

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

func SetupWorkload() string {
	rand.Seed(time.Now().UnixNano())
	name := fmt.Sprintf("test-workload-%v", rand.Intn(99999))
	args := []string{"-n", name, "nginx"}
	out, err := RioCmd("run", args)
	if err != nil {
		if err.Error() == "" {
			ginkgo.Fail(out)
		}
		ginkgo.Fail(err.Error())
	}
	err = WaitForWorkload(name)
	if err != nil {
		ginkgo.Fail(err.Error())
	}
	return name
}

func CleanupWorkload(name string) {
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

///////// Wait Methods

type waitForMe = func() bool

// WaitFor takes a method and waits until it returns true, see WaitForWorkload
func WaitFor(f waitForMe, timeout int) bool {
	sleepSeconds := 1
	for start := time.Now(); time.Since(start) < time.Second*time.Duration(timeout); {
		out := f()
		if out == true {
			return out
		}
		time.Sleep(time.Second * time.Duration(sleepSeconds))
		sleepSeconds++
	}
	return false
}

func WaitForWorkload(name string) error {
	f := func() bool {
		s, err := InspectService(name)
		if err == nil {
			if s.Status.DeploymentStatus != nil && s.Status.DeploymentStatus.AvailableReplicas > 0 {
				return true
			}
		}
		return false
	}
	ok := WaitFor(f, 120)
	if ok == false {
		return errors.New("workload failed to initiate")
	}
	return nil
}
