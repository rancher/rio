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
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/pkg/namesgenerator"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

const TestingNamespace = "testing-ns"

func init() {
	rand.Seed(time.Now().UnixNano())
}

type stop struct {
	error
}

// IntegrationPreCheck ensures CLI flag is passed, this way integration tests won't run during unit or validation tests
func IntegrationPreCheck() {
	runTests := flag.Bool("integration-tests", false, "must be provided to run the integration tests")
	flag.Parse()
	if !*runTests {
		fmt.Fprintln(os.Stderr, "integration test must be enabled with --integration-tests")
		os.Exit(0)
	}
}

// ValidationPreCheck ensures CLI flag is passed, this way validation tests won't run during unit or integration tests
func ValidationPreCheck() {
	runTests := flag.Bool("validation-tests", false, "must be provided to run the validation tests")
	flag.Parse()
	if !*runTests {
		fmt.Fprintln(os.Stderr, "validation test must be enabled with --validation-tests")
		os.Exit(0)
	}
}

// RioCmd executes rio CLI commands with your arguments in testing namespace
// Example: args=["run", "-n", "test", "nginx"] would run: "rio --namespace testing-namespace run -n test nginx"
func RioCmd(args []string, envs ...string) (string, error) {
	args = append([]string{"--namespace", TestingNamespace}, args...)
	cmd := exec.Command("rio", args...)
	cmd.Env = envs
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

// HeyCmd generates load on a specified URL
// Example: url=test-testing-ns.abcdef.on-rio.io, time=90s, c=120 would run: "hey -z 90s -c 120 http://test-testing-ns.abcdef.on-rio.io:9080"
func HeyCmd(url string, time string, c int) {
	args := []string{"-z", time, "-c", strconv.Itoa(c), url}
	cmd := exec.Command("hey", args...)
	stdOutErr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("%s: %s", err.Error(), stdOutErr))
	}
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
	err := wait.Poll(2*time.Second, 240*time.Second, f)
	if err != nil {
		return "", errors.New("endpoint did not return 200")
	}
	resp, _ := GetURL(endpoint)
	for i := 0; i < 5; i++ {
		if resp == "no healthy upstream" {
			time.Sleep(1 * time.Second)
			resp, _ = GetURL(endpoint)
		} else {
			break
		}
	}
	return resp, nil
}

// WaitForNoResponse waits until the response returned by a service is not 200
func WaitForNoResponse(endpoint string) (string, error) {
	f := wait.ConditionFunc(func() (bool, error) {
		_, err := GetURL(endpoint)
		if err == nil {
			return false, nil
		}
		return true, nil
	})
	err := wait.Poll(2*time.Second, 240*time.Second, f)
	if err != nil {
		return "", errors.New("endpoint did not go down")
	}
	resp, err := GetURL(endpoint)
	if err == nil {
		return resp, errors.New("endpoint did not go down")
	}
	return resp, nil

}

// GetURL performs an HTTP.Get on a endpoint and returns an error if the resp is not 200
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

func GetHostname(urls ...string) string {
	for _, u := range urls {
		u1, err := url.Parse(u)
		if err != nil {
			return ""
		}
		if u1.Scheme == "http" {
			return u
		}
	}

	return ""
}

// stringInSlice returns true if string a value is equals to any element of the slice otherwise false
func stringInSlice(a string, list []string) bool {
	for _, i := range list {
		if i == a {
			return true
		}
	}
	return false
}

// GetKubeClient returns the kubernetes clientset for querying its API, defaults to
// KUBECONFIG env value
func GetKubeClient() *kubernetes.Clientset {
	kubeConfigENV := os.Getenv("KUBECONFIG")
	if kubeConfigENV == "" {
		if home := homeDir(); home != "" {
			kubeConfigENV = filepath.Join(home, ".kube", "config")
		} else {
			fmt.Fprintln(os.Stderr, "an error occurred please set the KUBECONFIG environment variable")
			os.Exit(1)
		}
	}
	kubeConfig, _ := clientcmd.BuildConfigFromFlags("", kubeConfigENV)
	clientset, _ := kubernetes.NewForConfig(kubeConfig)
	return clientset
}

// homeDir returns the user HOME PATH
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

// CreateCNAME creates a CNAME if it doesn't exist or update if it exist already in the DNS Zone
// We use AWS Route53 as DNS provider.
func CreateCNAME(clusterDomain string) *route53.ChangeResourceRecordSetsOutput {
	sess, err := session.NewSession()
	if err != nil {
		fmt.Println("failed to create session make sure to add credentials to ~/.aws/credentials,", err)
		fmt.Println(err.Error())
	}
	cname := GetCNAMEInfo()
	zoneID := getZoneIDInfo()
	if cname == "" || clusterDomain == "" || zoneID == "" {
		fmt.Println(fmt.Errorf("incomplete information: d: %s, t: %s, z: %s", cname, clusterDomain, zoneID))
	}
	svc := route53.New(sess)

	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(cname),
						Type: aws.String("CNAME"),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(clusterDomain),
							},
						},
						TTL: aws.Int64(10),
					},
				},
			},
			Comment: aws.String("Add CNAME to Rio Cluster Domain"),
		},
		HostedZoneId: aws.String(zoneID),
	}
	resp, err := svc.ChangeResourceRecordSets(params)

	if err != nil {
		fmt.Println(err.Error())
	}
	return resp
}

// DeleteCNAME deletes a CNAME we provide as ENV var
// We use AWS Route53 as DNS provider.
func DeleteCNAME(clusterDomain string) *route53.ChangeResourceRecordSetsOutput {
	sess, err := session.NewSession()
	if err != nil {
		fmt.Println("failed to create session make sure to add credentials to ~/.aws/credentials,", err)
		fmt.Println(err.Error())
	}
	cname := GetCNAMEInfo()
	zoneID := getZoneIDInfo()
	if cname == "" || zoneID == "" {
		fmt.Println(fmt.Errorf("incomplete information: d: %s, t: %s", cname, zoneID))
	}
	svc := route53.New(sess)

	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("DELETE"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(cname),
						Type: aws.String("CNAME"),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(clusterDomain),
							},
						},
						TTL: aws.Int64(10),
					},
				},
			},
			Comment: aws.String("Delete CNAME to Rio Cluster Domain"),
		},
		HostedZoneId: aws.String(zoneID),
	}
	resp, err := svc.ChangeResourceRecordSets(params)

	if err != nil {
		fmt.Println(err.Error())
	}
	return resp
}

// GetCNAMEInfo retrieves the RIO_CNAME environment variable
func GetCNAMEInfo() string {
	return os.Getenv("RIO_CNAME")
}

// getZoneIDInfo retrieves the RIO_CNAME environment variable
func getZoneIDInfo() string {
	return os.Getenv("RIO_ROUTE53_ZONEID")
}
