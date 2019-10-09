package testutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

type TestRiofile struct {
	Name     string
	Filepath string
	Stack    riov1.Stack
	T        *testing.T
}

// Bring up a riofile by fixture file
func (trf *TestRiofile) Up(t *testing.T, filename string) {
	trf.T = t
	stackName := RandomString(5)
	trf.Name = fmt.Sprintf("%s/%s", testingNamespace, stackName)
	pwd, err := os.Getwd()
	if err != nil {
		trf.T.Fatal(err)
	}
	if !strings.Contains(pwd, "tests/") {
		pwd = pwd + "/tests/integration" // hack for running tests in package dir
	}
	trf.Filepath = fmt.Sprintf("%s/fixtures/%s", pwd, filename)
	_, err = RioCmd([]string{"--namespace", testingNamespace, "up", "--name", stackName, "-f", trf.Filepath})
	if err != nil {
		trf.T.Fatalf("Failed to create stack:  %v", err.Error())
	}
	err = trf.waitForStack()
	if err != nil {
		trf.T.Fatalf(err.Error())
	}
}

// Remove a stack and its objects
// todo: use owner-name annotation to remove orphaned objects (potentially in pkg) if we continue to see them
func (trf *TestRiofile) Remove() {
	_, err := RioCmd([]string{"rm", "--type", "stack", trf.Name})
	if err != nil {
		trf.T.Log(err.Error())
	}
}

// Return "rio export --stack {name}" as string
func (trf *TestRiofile) ExportStack() string {
	out, err := RioCmd([]string{"export", "--stack", trf.Name})
	if err != nil {
		trf.T.Log(err.Error())
	}
	return strings.TrimSuffix(out, "\n")
}

// Returns raw Riofile as string
func (trf *TestRiofile) Readfile() string {
	contents, err := ioutil.ReadFile(trf.Filepath)
	if err != nil {
		trf.T.Log(err.Error())
	}
	return strings.TrimSuffix(string(contents), "\n")
}

//////////////////
// Private methods
//////////////////

// reload calls inspect on the stack and uses that to reload our object
func (trf *TestRiofile) reload() error {
	out, err := RioCmd([]string{"inspect", "--type", "stack", "--format", "json", trf.Name})
	if err != nil {
		return err
	}
	trf.Stack = riov1.Stack{}
	if err := json.Unmarshal([]byte(out), &trf.Stack); err != nil {
		return err
	}
	return nil
}

// Wait for a stack to be deployed, does not wait for stack objects
func (trf *TestRiofile) waitForStack() error {
	f := wait.ConditionFunc(func() (bool, error) {
		err := trf.reload()
		if err == nil {
			for _, con := range trf.Stack.Status.Conditions {
				if con.Type == "DeployedStackRiofile" && con.Status == "True" {
					return true, nil
				}
			}
		}
		return false, nil
	})
	err := wait.Poll(2*time.Second, 60*time.Second, f)
	if err != nil {
		return errors.New("stack not successfully initiated")
	}
	return nil
}
