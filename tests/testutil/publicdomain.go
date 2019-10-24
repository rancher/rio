package testutil

import (
	"encoding/json"
	//"errors"
	"fmt"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	//"k8s.io/apimachinery/pkg/util/wait"
	"strings"
	"testing"
	//"time"
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
	if td.PublicDomain.Name != "" {
		_, err := RioCmd([]string{"domain", "unregister", td.GeneratedDomainName})
		if err != nil {
			td.T.Logf("failed to unregister domain:  %v", err.Error())
		}
	}
}

// GetDomain returns standard format non-namespaced domain, ex: "foo.bar"
func (td *TestDomain) GetDomain() string {
	err := td.reload()
	if err != nil {
		td.T.Fatalf("failed to fetch domain: %v", err.Error())
	}
	return getStandardFormatDomain(td.PublicDomain)
}

// GetKubeDomain receives the TestDomain object to retrieve the test PublicDomain data
// CLI Command Run: "kubectl get publicdomains my-domain -n testing-ns -o json"
func (td *TestDomain) GetKubeDomain() string {
	td.reload()
	args := []string{"get", "publicdomains", td.PublicDomain.GetName(), "-n", testingNamespace, "-o", "json"}
	resultString, err := KubectlCmd(args)
	if err != nil {
		td.T.Fatalf("Failed to get admin.rio.cattle.io.publicdomains:  %v", err.Error())
	}
	var results adminv1.PublicDomain
	err = json.Unmarshal([]byte(resultString), &results)
	if err != nil {
		td.T.Fatalf("Failed to unmarshal PublicDomain result: %s with error: %v", resultString, err.Error())
	}
	return getStandardFormatDomain(results)
}

//////////////////
// Private methods
//////////////////

func (td *TestDomain) reload() error {
	out, err := RioCmd([]string{"inspect", "--format", "json", td.GeneratedDomainName})
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
	//f := wait.ConditionFunc(func() (bool, error) {
	//	err := td.reload()
	//	if err == nil {
	//		td.PublicDomain.Status.Conditions
	//		if td.PublicDomain.Status.Endpoint != "" {
	//			return true, nil
	//		}
	//	}
	//	return false, nil
	//})
	//err := wait.Poll(2*time.Second, 60*time.Second, f)
	//if err != nil {
	//	return errors.New("domain not successfully created")
	//}
	return nil
}

// getStandardFormatDomain takes in a PublicDomain object
// Returns standard format non-namespaced public domain name, ex: "foo.bar"
func getStandardFormatDomain(publicDomain adminv1.PublicDomain) string {
	//if publicDomain.Spec.DomainName == "" {
	return ""
	//}
	//return strings.Replace(publicDomain.Spec.DomainName, "-", ".", 1)
}
