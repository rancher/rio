package testutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/onsi/ginkgo"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

const testingNamespace = "testing-ns"

func GetServiceName(name string) string {
	return fmt.Sprintf("%s/%s", testingNamespace, name)
}

func SetupService() string {
	rand.Seed(time.Now().UnixNano())
	name := fmt.Sprintf("test-service-%v", rand.Intn(99999))
	name = GetServiceName(name)
	out, err := RioCmd("run", []string{"-n", name, "nginx"})
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

func WaitForService(name string) error {
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
		return errors.New("service failed to initiate")
	}
	return nil
}
