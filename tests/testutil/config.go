package testutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

type TestConfig struct {
	Name      string
	Filepath  string
	ConfigMap corev1.ConfigMap
	T         *testing.T
}

// Executes "rio config create ns/randomconfig {fileWithContent}"
// This does not take a name or namespace param, that is setup by default
func (tc *TestConfig) Create(t *testing.T, content []string) {
	tc.T = t
	name := GenerateName()
	tc.Name = fmt.Sprintf("%s/%s", testingNamespace, name)
	err := tc.createTempFile(name, content)
	defer os.Remove(tc.Filepath)
	if err != nil {
		tc.T.Fatal(err.Error())
	}
	args := []string{"create", tc.Name, tc.Filepath}
	_, err = RioCmd("config", args)
	if err != nil {
		tc.T.Fatalf("config create command failed: %v", err.Error())
	}
	err = tc.waitForConfig()
	if err != nil {
		tc.T.Fatalf(err.Error())
	}
}

// Executes "rio rm" for this config
func (tc *TestConfig) Remove() {
	if tc.ConfigMap.Name != "" {
		_, err := RioCmd("rm", []string{"--type", "config", tc.Name})
		if err != nil {
			tc.T.Logf("failed to delete config: %v", err.Error())
		}
	}
}

// GetContent returns the configs Data.Content as list of strings, newline separated
func (tc *TestConfig) GetContent() []string {
	var data []string
	if val, ok := tc.ConfigMap.Data["content"]; ok {
		if val != "" {
			for _, s := range strings.Split(val, "\n") {
				if s != "" {
					data = append(data, s)
				}
			}
		}
	}
	return data
}

//////////////////
// Private methods
//////////////////

func (tc *TestConfig) createTempFile(filename string, content []string) error {
	tmpFile, err := ioutil.TempFile("", filename)
	if err != nil {
		tc.T.Fatal(err)
	}
	tc.Filepath = tmpFile.Name()
	for _, s := range content {
		if _, err := tmpFile.WriteString(s + "\n"); err != nil {
			return err
		}
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}
	return nil
}

func (tc *TestConfig) reload() error {
	args := append([]string{"--type", "config", "--format", "json", tc.Name})
	out, err := RioCmd("inspect", args)
	if err != nil {
		return err
	}
	tc.ConfigMap = corev1.ConfigMap{}
	if err := json.Unmarshal([]byte(out), &tc.ConfigMap); err != nil {
		return err
	}
	return nil
}

func (tc *TestConfig) waitForConfig() error {
	f := func() bool {
		err := tc.reload()
		if err == nil && tc.ConfigMap.UID != "" {
			return true
		}
		return false
	}
	ok := WaitFor(f, 60)
	if ok == false {
		return errors.New("config not successfully created")
	}
	return nil
}
