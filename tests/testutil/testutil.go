package testutil

import (
	"bufio"
	"context"
	cryptRand "crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
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
	return RioExecute(args, envs...)
}

// RioExecute executes rio CLI commands with arguments, use RioCmd unless you have to use a non-testing namespace
func RioExecute(args []string, envs ...string) (string, error) {
	cmd := exec.Command("rio", args...)
	cmd.Env = envs
	stdOutErr, err := cmd.CombinedOutput()
	if os.Getenv("DEBUG_TEST") != "" {
		fmt.Println(args)
		fmt.Println(envs)
	}
	if err != nil {
		return "", fmt.Errorf("%s: %s", err.Error(), stdOutErr)
	}
	return string(stdOutErr), nil
}

// RioExecuteWithRetry executes rio CLI commands with your arguments
// Example: args=["run", "-n", "test", "nginx"] would run: "rio run -n test nginx"
func RioExecuteWithRetry(args []string, envs ...string) (string, error) {
	out, err := retry(5, 1, RioExecute, args, envs...)
	if err != nil {
		return "", fmt.Errorf("%s: %s", err.Error(), out)
	}
	return out, nil
}

// RioCmdWithRetry executes rio CLI commands with your arguments
// Example: args=["run", "-n", "test", "nginx"] would run: "rio --namespace testing-namespace run -n test nginx"
func RioCmdWithRetry(args []string, envs ...string) (string, error) {
	out, err := retry(5, 1, RioCmd, args, envs...)
	if err != nil {
		return "", fmt.Errorf("%s: %s", err.Error(), out)
	}
	return out, nil
}

// RioCmdWithTail executes rio CLI commands that tail output with your arguments in testing namespace.
// Example: args=["attach", "nginx"] would run: "rio --namespace testing-namespace attach nginx"
func RioCmdWithTail(timeoutSeconds int, args []string, envs ...string) ([]string, error) {
	var results []string
	args = append([]string{"--namespace", TestingNamespace}, args...)
	timeout, _ := time.ParseDuration(fmt.Sprintf("%vs", timeoutSeconds))

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "rio", args...)
	out, _ := cmd.StdoutPipe()
	cmd.Start()
	scanner := bufio.NewScanner(out)
	go func() {
		for scanner.Scan() {
			if scanner.Text() != "" || cmd.ProcessState.Exited() {
				results = append(results, scanner.Text())
			}
		}
	}()
	cmd.Wait()

	return results, nil
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
	var globalErr error
	f := wait.ConditionFunc(func() (bool, error) {
		_, err := GetURL(endpoint)
		if err == nil {
			return true, nil
		}
		globalErr = err
		return false, nil
	})
	err := wait.Poll(2*time.Second, 240*time.Second, f)
	if err != nil {
		return "", fmt.Errorf("endpoint did not return 200, err: %v", globalErr)
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
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	response, err := client.Get(url)
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
		if u1.Scheme == "http" || u1.Scheme == "https" {
			return u1.Hostname()
		}
	}

	return strings.Join(urls, "")
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
func GetKubeClient(t *testing.T) *kubernetes.Clientset {
	kubeConfigENV := os.Getenv("KUBECONFIG")
	if kubeConfigENV == "" {
		if home := homeDir(); home != "" {
			kubeConfigENV = filepath.Join(home, ".kube", "config")
		} else {
			t.Fatalf("An error occurred when retrieving the kubectl client. Please set your KUBECONFIG environment variable.")
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

// CreateSelfSignedCert creates public and private keys for a self signed certificate
func CreateSelfSignedCert(t *testing.T) {
	key, err := rsa.GenerateKey(cryptRand.Reader, 2048)
	if err != nil {
		t.Fatalf("Private key cannot be created. %v\n", err.Error())
	}
	// Generate a pem block with the private key
	keyPemFile, _ := os.Create("testkey.pem")
	var pemKey = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	_ = pem.Encode(keyPemFile, pemKey)
	_ = keyPemFile.Close()

	tml := x509.Certificate{
		// you can add any attr that you need
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(1, 0, 0),
		// you have to generate a different serial number each execution
		SerialNumber: big.NewInt(123123),
		Subject: pkix.Name{
			CommonName:   "Rio Test Automation",
			Organization: []string{"Rancher - Rio QA"},
		},
		BasicConstraintsValid: true,
	}
	cert, err := x509.CreateCertificate(cryptRand.Reader, &tml, &tml, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("Certificate cannot be created. %v\n", err.Error())
	}
	// Generate a pem block with the certificate
	certPemFile, _ := os.Create("testcert.pem")
	var pemCert = &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	}
	_ = pem.Encode(certPemFile, pemCert)
	_ = certPemFile.Close()
}

// DeleteSelfSignedCert deletes the generated cert files
func DeleteSelfSignedCert() {
	_ = os.Remove("testkey.pem")
	_ = os.Remove("testcert.pem")
}

// CreateRoute53DNS creates a CNAME if it doesn't exist or update if it exist already in the DNS Zone
// We use AWS Route53 as DNS provider.
func CreateRoute53DNS(t *testing.T, recordType, clusterDomain string) *route53.ChangeResourceRecordSetsOutput {
	params := SetRoute53Params(t, recordType, clusterDomain)
	sess, err := session.NewSession()
	if err != nil {
		t.Fatalf("Failed to create session. Make sure to add credentials to ~/.aws/credentials or pass in required environment variables %v\n", err.Error())
	}
	svc := route53.New(sess)
	resp, err := svc.ChangeResourceRecordSets(params)
	if err != nil {
		t.Fatalf("Failed to create Route53 DNS: %v\n", err.Error())
	}
	return resp
}

// SetRoute53Params creates a CNAME if it doesn't exist or update if it exist already in the DNS Zone
// We use AWS Route53 as DNS provider.
func SetRoute53Params(t *testing.T, recordType, ipOrDNS string) *route53.ChangeResourceRecordSetsInput {
	var recordName string
	if recordType == "A" {
		recordName = "*." + GetAInfo()
	} else if recordType == "CNAME" {
		recordName = GetCNAMEInfo()
	}
	zoneID := getZoneIDInfo()
	if recordName == "" || ipOrDNS == "" || zoneID == "" {
		t.Fatal("incomplete information supplied to create a Route53 record.")
	}

	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(recordName),
						Type: aws.String(recordType),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(ipOrDNS),
							},
						},
						TTL: aws.Int64(10),
					},
				},
			},
			Comment: aws.String("DNS for custom record in Rio Cluster Domains"),
		},
		HostedZoneId: aws.String(zoneID),
	}
	return params
}

// DeleteRoute53DNS deletes a CNAME we provide as ENV var
// We use AWS Route53 as DNS provider.
func DeleteRoute53DNS(clusterDomain string) *route53.ChangeResourceRecordSetsOutput {
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
	if os.Getenv("RIO_CNAME") == "" {
		return "riotestautocname." + os.Getenv("RIO_ROUTE53_ZONENAME")
	}
	return os.Getenv("RIO_CNAME")
}

// GetAInfo retrieves the RIO_A_RECORD environment variable
func GetAInfo() string {
	if os.Getenv("RIO_A_RECORD") == "" {
		return "riotestautoa." + os.Getenv("RIO_ROUTE53_ZONENAME")
	}
	return os.Getenv("RIO_A_RECORD")
}

// getZoneIDInfo retrieves the RIO_CNAME environment variable
func getZoneIDInfo() string {
	return os.Getenv("RIO_ROUTE53_ZONEID")
}

// retry function is intended for retrying command line commands invocations that collisioned while updating kubernetes objects
func retry(attempts int, sleep time.Duration, f func(args []string, envs ...string) (string, error), args []string, envs ...string) (string, error) {
	strOut, err := f(args, envs...)
	if err != nil {
		if s, ok := err.(stop); ok {
			return strOut, s.error
		}
		if attempts--; attempts > 0 {
			jitter := time.Duration(rand.Int63n(int64(sleep)))
			sleep = sleep + jitter/2
			time.Sleep(sleep)
			return retry(attempts, 2*sleep, f, args, envs...)
		}
		return strOut, err
	}
	return strOut, nil
}
