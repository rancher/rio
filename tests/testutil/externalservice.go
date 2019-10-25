package testutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

type TestExternalService struct {
	Target          string
	Name            string
	ExternalService riov1.ExternalService
	T               *testing.T
}

// Executes "rio externalservice create ns/randomservice {target}"
// This does not take a name or namespace param, that is setup by default
func (es *TestExternalService) Create(t *testing.T, target string) {
	es.T = t
	es.Target = target
	name := GenerateName()
	es.Name = fmt.Sprintf("%s/%s", "externalservice", name)
	_, err := RioCmd([]string{"externalservice", "create", name, target})
	if err != nil {
		es.T.Fatalf("external service create command failed: %v", err.Error())
	}
	err = es.waitForExternalService()
	if err != nil {
		es.T.Fatalf(err.Error())
	}
}

// Takes the name of an existing external service, loads it, and returns
func GetExternalService(t *testing.T, name string) TestExternalService {
	es := TestExternalService{
		Target:          "",
		Name:            fmt.Sprintf("externalservice/%s/%s", testingNamespace, name),
		ExternalService: riov1.ExternalService{},
		T:               t,
	}
	err := es.waitForExternalService()
	if err != nil {
		es.T.Fatalf(err.Error())
	}
	return es
}

// Executes "rio rm" for this external service
func (es *TestExternalService) Remove() {
	if es.ExternalService.Name != "" {
		_, err := RioCmd([]string{"rm", es.Name})
		if err != nil {
			es.T.Logf("failed to delete external service: %v", err.Error())
		}
	}
}

// There can be multiple IPAddresses on a service, this returns first
func (es *TestExternalService) GetFirstIPAddress() string {
	if len(es.ExternalService.Spec.IPAddresses) == 0 {
		return ""
	}
	return es.ExternalService.Spec.IPAddresses[0]
}

func (es *TestExternalService) GetFQDN() string {
	return es.ExternalService.Spec.FQDN
}

//////////////////
// Private methods
//////////////////

func (es *TestExternalService) reload() error {
	out, err := RioCmd([]string{"inspect", "--format", "json", es.Name})
	if err != nil {
		return err
	}
	es.ExternalService = riov1.ExternalService{}
	if err := json.Unmarshal([]byte(out), &es.ExternalService); err != nil {
		return err
	}
	return nil
}

func (es *TestExternalService) waitForExternalService() error {
	f := wait.ConditionFunc(func() (bool, error) {
		err := es.reload()
		if err == nil && (len(es.ExternalService.Spec.IPAddresses) > 0 || es.ExternalService.Spec.FQDN != "") {
			return true, nil
		}
		return false, nil
	})
	err := wait.Poll(2*time.Second, 30*time.Second, f)
	if err != nil {
		return errors.New("external service not successfully initiated")
	}
	return nil
}
