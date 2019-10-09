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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
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
	_, err = RioCmd([]string{"config", "create", tc.Name, tc.Filepath})
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
		_, err := RioCmd([]string{"rm", "--type", "config", tc.Name})
		if err != nil {
			tc.T.Logf("failed to delete config: %v", err.Error())
		}
	}
}

// GetContent returns the configs Data.Content as list of strings, newline separated
func (tc *TestConfig) GetContent() []string {
	return getContentData(tc.ConfigMap)
}

// GetKubeContent returns the kubectl configmap's Data.Content as list of strings, newline separated
// CLI Command Run: kubectl get cm testname -n testing-ns -o json
func (tc *TestConfig) GetKubeContent() []string {
	args := []string{"get", "cm", tc.ConfigMap.Name, "-n", testingNamespace, "-o", "json"}
	resultString, err := KubectlCmd(args)
	if err != nil {
		tc.T.Fatalf("Failed to get ConfigMaps:  %v", err.Error())
	}
	var results corev1.ConfigMap
	err = json.Unmarshal([]byte(resultString), &results)
	if err != nil {
		tc.T.Fatalf("Failed to unmarshal ConfigMaps result: %s with error: %v", resultString, err.Error())
	}

	return getContentData(results)
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
	out, err := RioCmd([]string{"inspect", "--type", "config", "--format", "json", tc.Name})
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
	f := wait.ConditionFunc(func() (bool, error) {
		err := tc.reload()
		if err == nil && tc.ConfigMap.UID != "" {
			return true, nil
		}
		return false, nil
	})
	err := wait.Poll(2*time.Second, 60*time.Second, f)
	if err != nil {
		return errors.New("config not successfully created")
	}
	return nil
}

// getContentData returns the configs Data.Content as list of strings, newline separated
func getContentData(cm corev1.ConfigMap) []string {
	var data []string
	if val, ok := cm.Data["content"]; ok {
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
