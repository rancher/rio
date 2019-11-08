package testutil

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/yaml"
)

type TestRiofile struct {
	Name       string
	StackName  string
	Filepath   string
	Stack      riov1.Stack
	T          *testing.T
	Kubeconfig string
}

// Bring up a riofile by fixture file
func (trf *TestRiofile) Up(t *testing.T, filename, stackName string, args ...string) {
	trf.create(t, filename, stackName)
	var envs []string
	if trf.Kubeconfig != "" {
		envs = []string{fmt.Sprintf("KUBECONFIG=%s", trf.Kubeconfig)}
	}
	_, err := RioCmd(append([]string{"up", "--name", trf.StackName, "-f", trf.Filepath}, args...), envs...)
	if err != nil {
		trf.T.Fatalf("Failed to create stack:  %v", err.Error())
	}
}

func (trf *TestRiofile) UpWithRepo(t *testing.T, repoName, stackName string, args ...string) error {
	trf.create(t, "", stackName)
	var envs []string
	if trf.Kubeconfig != "" {
		envs = []string{fmt.Sprintf("KUBECONFIG=%s", trf.Kubeconfig)}
	}
	upArgs := append([]string{"up", "--name", trf.StackName}, args...)
	_, err := RioCmd(append(upArgs, repoName), envs...)
	if err != nil {
		return err
	}
	return nil
}

func (trf *TestRiofile) create(t *testing.T, filename, stackName string) {
	trf.T = t
	if stackName == "" {
		stackName = RandomString(5)
	}
	trf.StackName = stackName
	trf.Name = fmt.Sprintf("%s:%s/%s", TestingNamespace, "stack", stackName)
	pwd, err := os.Getwd()
	if err != nil {
		trf.T.Fatal(err)
	}
	if !strings.Contains(pwd, "tests/") {
		pwd = pwd + "/tests/integration" // hack for running tests in package dir
	}
	trf.Filepath = fmt.Sprintf("%s/fixtures/%s", pwd, filename)
}

// Remove a stack and its objects
// todo: use owner-name annotation to remove orphaned objects (potentially in pkg) if we continue to see them
func (trf *TestRiofile) Remove() {
	if trf.T == nil {
		return
	}
	_, err := RioCmd([]string{"rm", trf.Name})
	if err != nil {
		trf.T.Log(err.Error())
	}
}

// Return "rio export --stack {name}"
func (trf *TestRiofile) ExportStack() (map[string]interface{}, error) {
	content, err := RioCmd([]string{"export", "--riofile", trf.Name})
	if err != nil {
		trf.T.Log(err.Error())
	}
	data := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(content), &data); err != nil {
		return nil, err
	}
	return data, nil
}

// Returns raw Riofile
func (trf *TestRiofile) Readfile() (map[string]interface{}, error) {
	content, err := ioutil.ReadFile(trf.Filepath)
	if err != nil {
		trf.T.Log(err.Error())
	}
	data := map[string]interface{}{}
	if err := yaml.Unmarshal(content, &data); err != nil {
		return nil, err
	}
	return data, nil
}

//////////////////
// Private methods
//////////////////

// reload calls inspect on the stack and uses that to reload our object
func (trf *TestRiofile) reload() error {
	out, err := RioCmd([]string{"inspect", "--format", "json", trf.Name})
	if err != nil {
		return err
	}
	trf.Stack = riov1.Stack{}
	if err := json.Unmarshal([]byte(out), &trf.Stack); err != nil {
		return err
	}
	return nil
}
