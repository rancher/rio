package testutil

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/docker/docker/pkg/namesgenerator"
	"k8s.io/apimachinery/pkg/util/wait"
)

const testingNamespace = "testing-ns"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Ensure CLI flag is passed, this way integration tests won't run during unit tests
func PreCheck() {
	runTests := flag.Bool("integration-tests", false, "must be provided to run the integration tests")
	flag.Parse()
	if !*runTests {
		fmt.Fprintln(os.Stderr, "integration test must be enabled with --integration-tests")
		os.Exit(0)
	}
}

// RioCmd executes rio CLI commands with your arguments
// Example: args=["run", "-n", "test", "nginx"] would run: "rio run -n test nginx"
func RioCmd(args []string) (string, error) {
	cmd := exec.Command("rio", args...)
	stdOutErr, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %s", err.Error(), stdOutErr)
	}
	return string(stdOutErr), nil
}

// KubectlCmd executes kubectl CLI commands with your arguments
// Example: args=["get", "-n", "test", "services"] would run: "kubectl get -n test services"
func KubectlCmd(args []string) (string, error) {
	cmd := exec.Command("kubectl", args...)
	stdOutErr, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %s", err.Error(), stdOutErr)
	}
	return string(stdOutErr), nil
}

// Wait until a URL has a response that returns 200 status code, else return error
func WaitForURLResponse(endpoint string) (string, error) {
	f := wait.ConditionFunc(func() (bool, error) {
		_, err := GetURL(endpoint)
		if err == nil {
			return true, nil
		}
		return false, nil
	})
	err := wait.Poll(2*time.Second, 120*time.Second, f)
	if err != nil {
		return "", errors.New("endpoint did not return 200")
	}
	resp, _ := GetURL(endpoint)
	return resp, nil
}

func GetURL(url string) (string, error) {
	var body string
	response, err := http.Get(url)
	if err != nil {
		return body, err
	}
	defer response.Body.Close()
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return body, err
	}
	body = string(bytes)
	body = strings.TrimSuffix(body, "\n")
	if response.StatusCode != http.StatusOK {
		return body, fmt.Errorf("%s returned %d - %s", url, response.StatusCode, body)
	}
	return body, nil
}

// Return a random set of lowercase letters
func RandomString(length int) string {
	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		bytes[i] = byte(97 + rand.Intn(122-97))
	}
	return string(bytes)
}

func GenerateName() string {
	return strings.Replace(namesgenerator.GetRandomName(2), "_", "-", -1)
}

func GetHostname(URL string) string {
	u, err := url.Parse(URL)
	if err != nil {
		return ""
	}
	return u.Hostname()
}
