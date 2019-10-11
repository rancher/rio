package testutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
)

type TestDomain struct {
	GeneratedDomainName string
	PublicDomain        adminv1.PublicDomain
	T                   *testing.T
}

// Generates and returns a random string to use as domain name, ex: qpwb.towv
func GenerateRandomDomain() string {
	return fmt.Sprintf("%v.%v", RandomString(4), RandomString(4))
}

// Executes "rio domain register {domain} {target}" and returns a TestDomain
func (td *TestDomain) RegisterDomain(t *testing.T, domain string, target string) {
	td.T = t
	td.GeneratedDomainName = fmt.Sprintf("%v/%v",
		testingNamespace,
		strings.Replace(domain, ".", "-", 1))
	_, err := RioCmd([]string{"domain", "register", domain, target})
	if err != nil {
		td.T.Fatalf("register domain command failed: %v", err.Error())
	}
	err = td.waitForDomain()
	if err != nil {
		td.T.Fatalf(err.Error())
	}
}

// Executes "rio domain unregister" for this domain
func (td *TestDomain) UnRegister() {
	if td.PublicDomain.Spec.DomainName != "" {
		_, err := RioCmd([]string{"domain", "unregister", td.GeneratedDomainName})
		if err != nil {
			td.T.Logf("failed to unregister domain:  %v", err.Error())
		}
	}
}

// GetDomainName returns standard format non-namespaced domain, ex: "foo.bar"
func (td *TestDomain) GetDomain() string {
	err := td.reload()
	if err != nil {
		td.T.Fatalf("failed to fetch domain: %v", err.Error())
	}
	if td.PublicDomain.Spec.DomainName == "" {
		return ""
	}
	return strings.Replace(td.PublicDomain.Spec.DomainName, "-", ".", 1)
}

//////////////////
// Private methods
//////////////////

func (td *TestDomain) reload() error {
	out, err := RioCmd([]string{"inspect", "--type", "publicdomain", "--format", "json", td.GeneratedDomainName})
	if err != nil {
		return err
	}
	td.PublicDomain = adminv1.PublicDomain{}
	if err := json.Unmarshal([]byte(out), &td.PublicDomain); err != nil {
		return err
	}
	return nil
}

func (td *TestDomain) waitForDomain() error {
	f := wait.ConditionFunc(func() (bool, error) {
		err := td.reload()
		if err == nil {
			if td.PublicDomain.Status.Endpoint != "" {
				return true, nil
			}
		}
		return false, nil
	})
	err := wait.Poll(2*time.Second, 60*time.Second, f)
	if err != nil {
		return errors.New("domain not successfully created")
	}
	return nil
}
