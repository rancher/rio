package testutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/knative/pkg/apis"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

type TestService struct {
	Name    string // namespace/name:version
	AppName string // namespace/name
	App     riov1.App
	Service riov1.Service
	Build   tektonv1alpha1.TaskRun
	Version string
	T       *testing.T
}

// Create generates a new rio service, named randomly in the testing namespace, and
// returns a new TestService with it attached. Guarantees ready state but not live endpoint
func (ts *TestService) Create(t *testing.T, source ...string) {
	args := ts.createArgs(t, source...)
	_, err := RioCmd(args)
	if err != nil {
		ts.T.Fatalf("Failed to create service %s: %v", ts.Name, err.Error())
	}
	if ts.isGithubSource(source...) {
		err = ts.waitForBuild()
		if err != nil {
			ts.T.Fatalf(err.Error())
		}
	}
	err = ts.waitForReadyService()
	if err != nil {
		ts.T.Fatalf(err.Error())
	}
	err = ts.waitForAvailableReplicas(ts.GetScale())
	if err != nil {
		ts.T.Fatalf(err.Error())
	}
}

func (ts *TestService) createArgs(t *testing.T, source ...string) []string {
	ts.T = t
	ts.Version = "v0"
	ts.AppName = fmt.Sprintf(
		"%s/%s",
		testingNamespace,
		RandomString(5),
	)
	ts.Name = fmt.Sprintf("%s:%s", ts.AppName, ts.Version)
	if len(source) == 0 {
		source = []string{"nginx"}
	}
	// Ensure port 80/http unless the source is from github
	if !ts.isGithubSource(source...) {
		source = append([]string{"-p", "80/http"}, source...)
	}
	args := append([]string{"run", "-n", ts.AppName}, source...)
	return args
}

func (ts *TestService) isGithubSource(source ...string) bool {
	return len(source) > 0 && source[len(source)-1][0:4] == "http" && strings.Contains(source[len(source)-1], "github")
}

// Takes name and version of existing service and returns loaded TestService
func GetService(t *testing.T, name string, version string) TestService {
	ts := TestService{
		App:     riov1.App{},
		Service: riov1.Service{},
		Version: version,
		T:       t,
	}
	ts.AppName = fmt.Sprintf(
		"%s/%s",
		testingNamespace,
		name,
	)
	ts.Name = fmt.Sprintf("%s:%s", ts.AppName, ts.Version)
	err := ts.waitForReadyService()
	if err != nil {
		ts.T.Fatalf(err.Error())
	}
	return ts
}

// Remove calls "rio rm" on this service. Logs error but does not fail test.
func (ts *TestService) Remove() {
	if ts.Service.Status.DeploymentStatus != nil {
		_, err := RioCmd([]string{"rm", "--type", "service", ts.Name})
		if err != nil {
			ts.T.Log(err.Error())
		}
	}
}

// Call "rio scale ns/service={scaleTo}"
func (ts *TestService) Scale(scaleTo int) {
	_, err := RioCmd([]string{"scale", fmt.Sprintf("%s=%d", ts.Name, scaleTo)})
	if err != nil {
		ts.T.Fatalf("scale command failed:  %v", err.Error())
	}
	// First wait for scale on service to update
	err = ts.waitForScale(scaleTo)
	if err != nil {
		ts.T.Fatal(err.Error())
	}
	// Then wait for actual replicas to come up
	err = ts.waitForAvailableReplicas(scaleTo)
	if err != nil {
		ts.T.Fatal(err.Error())
	}
}

// Weight calls "rio weight --rollout={rollout} --rollout-increment={increment} --rollout-interval={interval} ns/service:version={percentage}" on this service.
// If passing rollout=false then the increment and interval values won't matter. Best practice is to pass "5" for both to keep in line with Spec defaults.
func (ts *TestService) Weight(percentage int, rollout bool, increment int, interval int) {
	_, err := RioCmd([]string{
		"weight",
		fmt.Sprintf("--rollout=%t", rollout),
		fmt.Sprintf("--rollout-increment=%d", increment),
		fmt.Sprintf("--rollout-interval=%d", interval),
		fmt.Sprintf("%s=%d", ts.Name, percentage),
	})
	if err != nil {
		ts.T.Fatalf("weight command failed:  %v", err.Error())
	}
	if !rollout {
		err = ts.waitForWeight(percentage)
		if err != nil {
			ts.T.Fatal(err.Error())
		}
	}
}

// Call "rio stage --image={source} ns/name:{version}", this will return a new TestService
func (ts *TestService) Stage(source, version string) TestService {
	name := fmt.Sprintf("%s:%s", ts.AppName, version)
	_, err := RioCmd([]string{"stage", "--image", source, name})
	if err != nil {
		ts.T.Fatalf("stage command failed:  %v", err.Error())
	}
	stagedService := TestService{
		T:       ts.T,
		AppName: ts.AppName,
		Name:    name,
		Version: version,
	}
	err = stagedService.waitForReadyService()
	if err != nil {
		ts.T.Fatalf(err.Error())
	}
	return stagedService
}

// Promote calls "rio promote --rollout=false [args] ns/name:{version}" to instantly promote a revision
func (ts *TestService) Promote(args ...string) {
	args = append(
		[]string{"promote", "--rollout=false"},
		append(args, ts.Name)...)
	_, err := RioCmd(args)
	if err != nil {
		ts.T.Fatalf("stage command failed:  %v", err.Error())
	}
	err = ts.waitForWeight(100)
	if err != nil {
		ts.T.Fatalf(err.Error())
	}
}

// Logs calls "rio logs ns/service" on this service
func (ts *TestService) Logs(args ...string) string {
	args = append([]string{"logs"}, append(args, ts.AppName)...)
	out, err := RioCmd(args)
	if err != nil {
		ts.T.Fatalf("logs command failed:  %v", err.Error())
	}
	return out
}

// Exec calls "rio exec ns/service {command}" on this service
func (ts *TestService) Exec(command ...string) string {
	out, err := RioCmd(append([]string{"exec", ts.AppName}, command...))
	if err != nil {
		ts.T.Fatalf("exec command failed:  %v", err.Error())
	}
	return out
}

// Export calls "rio export {serviceName}" and returns that in a new TestService object
func (ts *TestService) Export() TestService {
	args := []string{"export", "--type", "service", "--format", "json", ts.Name}
	service, err := ts.loadExport(args)
	if err != nil {
		ts.T.Fatal(err.Error())
	}
	return service
}

// ExportRaw works the same as export, but with --raw flag
func (ts *TestService) ExportRaw() TestService {
	args := []string{"export", "--raw", "--type", "service", "--format", "json", ts.Name}
	service, err := ts.loadExport(args)
	if err != nil {
		ts.T.Fatal(err.Error())
	}
	return service
}

//////////
// Getters
//////////

// GetEndpointResponse performs an http.get against the service endpoint and returns response if
// status code is 200, otherwise it errors out
func (ts *TestService) GetEndpointResponse() string {
	response, err := WaitForURLResponse(ts.GetEndpointURL())
	if err != nil {
		ts.T.Fatal(err.Error())
	}
	return response
}

func (ts *TestService) GetAppEndpointResponse() string {
	response, err := WaitForURLResponse(ts.GetAppEndpointURL())
	if err != nil {
		ts.T.Fatal(err.Error())
	}
	err = ts.waitForReadyService() // After hitting the endpoint, the service should be ready again, so verify that it is
	if err != nil {
		ts.T.Fatal(err.Error())
	}
	return response
}

// Returns count of ready and available pods
func (ts *TestService) GetAvailableReplicas() int {
	if ts.Service.Status.DeploymentStatus != nil {
		return int(ts.Service.Status.DeploymentStatus.AvailableReplicas)
	}
	return 0
}

// Returns desired scale, different from current available replicas
func (ts *TestService) GetScale() int {
	if ts.Service.Spec.Scale != nil {
		return *ts.Service.Spec.Scale
	}
	return 0
}

// Return service's goal weight, this is different from weight service is currently at
func (ts *TestService) GetSpecWeight() int {
	return ts.Service.Spec.Weight
}

// Return service's actual current weight, not the spec (end-goal) weight
func (ts *TestService) GetCurrentWeight() int {
	ts.reloadApp()
	return getRevisionWeight(ts.App, ts.Version)
}

func (ts *TestService) GetImage() string {
	return ts.Service.Spec.Image
}

// GetRunningPods returns the kubectl overview of all running pods for this service's app in an array
// Each value in the array is a string, separated by spaces, that will have the Pod's NAME  READY  STATUS  RESTARTS  AGE in that order.
func (ts *TestService) GetRunningPods() []string {
	ts.reloadApp()
	args := append([]string{"get", "pods",
		"-n", testingNamespace,
		"-l", fmt.Sprintf("app=%s", ts.App.Name),
		"--field-selector", "status.phase=Running",
		"--no-headers"})
	out, err := KubectlCmd(args)
	if err != nil {
		ts.T.Fatalf("Failed to get running pods:  %v", err.Error())
	}

	kubeResult := strings.Split(strings.TrimSpace(out), "\n")
	runningPods := kubeResult[:0]
	for _, pod := range kubeResult {
		if strings.Contains(pod, "Running") {
			runningPods = append(runningPods, pod)
		}
	}
	return runningPods
}

// GetResponseCounts takes an array of expected response strings and sends numRequests requests to the service's app endpoint.
// If it gets a response other than one in the specified array, it throws a failure. Otherwise it returns individual counts of each response.
func (ts *TestService) GetResponseCounts(responses []string, numRequests int) map[string]int {
	var responseCounts = map[string]int{}
	var response string
	for i := 0; i < numRequests; i++ {
		response = ts.GetAppEndpointResponse()
		gotExpectedResponse := false
		for _, resp := range responses {
			if response == resp {
				responseCounts[resp]++
				gotExpectedResponse = true
				break
			}
		}
		if !gotExpectedResponse {
			ts.T.Fatalf("Failed to get one of the expected responses. Got: %v", response)
		}
	}
	return responseCounts
}

// GenerateLoad queries the endpoint multiple times in order to put load on the service.
// It will execute for up to 60 seconds until there is an Observed Scale on the service or the the AvailableReplicas equal the MaxScale
func (ts *TestService) GenerateLoad() {
	f := wait.ConditionFunc(func() (bool, error) {
		HeyCmd(GetHostname(ts.GetAppEndpointURL()), "5s", 10*(*ts.Service.Spec.MaxScale))
		ts.reloadApp()
		ts.reload()
		if ts.Service.Status.ObservedScale != nil || ts.GetAvailableReplicas() == *ts.Service.Spec.MaxScale {
			return true, nil
		}
		return false, nil
	})
	wait.Poll(5*time.Second, 60*time.Second, f)
	if ts.Service.Status.ObservedScale != nil {
		ts.waitForAvailableReplicas(*ts.Service.Status.ObservedScale)
	}
}

// GetEndpointURL returns the URL for this service's app
func (ts *TestService) GetEndpointURL() string {
	url, err := ts.waitForEndpointDNS()
	if err != nil {
		ts.T.Fatalf("Failed to get the endpoint url:  %v", err.Error())
		return ""
	}
	return url
}

// GetAppEndpointURL retrieves the service's app endpoint URL and returns it as string
func (ts *TestService) GetAppEndpointURL() string {
	url, err := ts.waitForAppEndpointDNS()
	if err != nil {
		ts.T.Fatalf("Failed to get the endpoint url:  %v", err.Error())
		return ""
	}
	return url
}

// GetKubeEndpointURL returns the app revision endpoint URL
// and returns it as string
func (ts *TestService) GetKubeEndpointURL() string {
	_, err := ts.waitForEndpointDNS()
	if err != nil {
		ts.T.Fatalf("Failed waiting for DNS:  %v", err.Error())
		return ""
	}
	args := []string{"get", "service.rio.cattle.io",
		"-n", testingNamespace,
		ts.Service.Name,
		"-o", `jsonpath="{.status.endpoints[0]}"`}
	url, err := KubectlCmd(args)
	if err != nil {
		ts.T.Fatalf("Failed to get endpoint url:  %v", err.Error())
		return ""
	}
	return strings.Replace(url, "\"", "", -1) // remove double quotes from output
}

// GetKubeAppEndpointURL returns the endpoint URL of the service's app
// by using kubectl and returns it as string
func (ts *TestService) GetKubeAppEndpointURL() string {
	_, err := ts.waitForAppEndpointDNS()
	if err != nil {
		ts.T.Fatalf("Failed waiting for DNS:  %v", err.Error())
		return ""
	}
	appName := strings.Split(ts.AppName, "/")[1]
	args := []string{"get", "apps",
		"-n", testingNamespace,
		appName,
		"-o", `jsonpath="{.status.endpoints[0]}"`}
	url, err := KubectlCmd(args)
	if err != nil {
		ts.T.Fatalf("Failed to get app endpoint url:  %v", err.Error())
		return ""
	}

	return strings.Replace(url, "\"", "", -1) // remove double quotes from output
}

// GetKubeCurrentWeight takes in a revision value and retrieves the actual current weight, not the spec (end-goal) weight
func (ts *TestService) GetKubeCurrentWeight() int {
	ts.reloadApp()
	args := []string{"get", "apps", ts.App.GetName(), "-n", testingNamespace, "-o", "json"}
	resultString, err := KubectlCmd(args)
	if err != nil {
		ts.T.Fatalf("Failed to get rio.cattle.io.apps:  %v", err.Error())
	}
	var app riov1.App
	err = json.Unmarshal([]byte(resultString), &app)
	if err != nil {
		ts.T.Fatalf("Failed to unmarshal App results: %s with error: %v", resultString, err.Error())
	}

	return getRevisionWeight(app, ts.Version)
}

// PodsResponsesMatchAvailableReplicas does a GetURL in the App endpoint and stores the response in a slice
// the length of the resulting slice should represent the number of responsive pods in a service.
// Returns true if the number of replicas is equal to the length of the responses slice.
func (ts *TestService) PodsResponsesMatchAvailableReplicas(path string, numberOfReplicas int) bool {
	i := 0
	replicasTimesRequests := numberOfReplicas * 8
	responses := make([]string, 0)
	for i < replicasTimesRequests {
		response, err := WaitForURLResponse(ts.GetAppEndpointURL() + path)
		if err != nil {
			ts.T.Fatal(err.Error())
		}
		i++
		if !stringInSlice(response, responses) {
			responses = append(responses, response)
		}
	}
	return len(responses) == numberOfReplicas
}

// KubeCompareReplicasValues get the app number of ready replicasets with a clientset
// and returns true if that value match the scale given
func (ts *TestService) GetKubeAvailableReplicas() int {
	clientset := GetKubeClient()
	replicaSetList, err := clientset.AppsV1().
		ReplicaSets(testingNamespace).
		List(metav1.ListOptions{LabelSelector: "app=" + ts.Service.Name})
	if err != nil {
		ts.T.Fatalf(err.Error())
	}
	if len(replicaSetList.Items) == 0 {
		return 0
	}
	return int(replicaSetList.Items[0].Status.ReadyReplicas)
}

// WaitForScaleDown waits until either 5 minutes pass or a service has scaled down to its minimum value
func (ts *TestService) WaitForScaleDown() error {
	f := wait.ConditionFunc(func() (bool, error) {
		err := ts.reload()
		if err == nil {
			if *ts.Service.Spec.MinScale == ts.GetAvailableReplicas() {
				return true, nil
			}
		}
		return false, nil
	})
	err := wait.Poll(5*time.Second, 300*time.Second, f)
	if err != nil {
		return fmt.Errorf("service %v failed to scale down after 5 minutes", ts.Name)
	}
	return nil
}

//////////////////
// Private methods
//////////////////

// reload calls inspect on the service and uses that to reload our object
func (ts *TestService) reload() error {
	out, err := RioCmd([]string{"inspect", "--type", "service", "--format", "json", ts.Name})
	if err != nil {
		return err
	}
	ts.Service = riov1.Service{}
	if err := json.Unmarshal([]byte(out), &ts.Service); err != nil {
		return err
	}
	return nil
}

// reloadBuild calls inspect on the build and uses that to reload our object
func (ts *TestService) reloadBuild() error {
	ts.reload()
	out, err := KubectlCmd([]string{"get", "taskrun", "-n", testingNamespace, "-l", "service-name=" + ts.Service.GetName(), "-o", "json"})
	if err != nil {
		return err
	}
	list := tektonv1alpha1.TaskRunList{}
	if err := json.Unmarshal([]byte(out), &list); err != nil {
		return err
	}
	if len(list.Items) > 0 {
		ts.Build = list.Items[0]
	}
	return nil
}

// reload calls inspect on the service's app and uses that to reload the app obj
func (ts *TestService) reloadApp() error {
	out, err := RioCmd([]string{"inspect", "--type", "app", "--format", "json", ts.AppName})
	if err != nil {
		return err
	}
	ts.App = riov1.App{}
	if err := json.Unmarshal([]byte(out), &ts.App); err != nil {
		return err
	}
	return nil
}

// load a "rio export..." response into a new TestService obj
func (ts *TestService) loadExport(args []string) (TestService, error) {
	out, err := RioCmd(args)
	if err != nil {
		return TestService{}, fmt.Errorf("export command failed: %v", err.Error())
	}
	exportedService := riov1.Service{}
	if err := json.Unmarshal([]byte(out), &exportedService); err != nil {
		return TestService{}, fmt.Errorf("failed to parse export data: %v", err.Error())
	}
	stagedService := TestService{
		Service: exportedService,
		AppName: ts.AppName,
		Name:    ts.Name,
		Version: ts.Version,
	}
	return stagedService, nil
}

// Wait longer for higher scale, always wait at least 120 seconds
func (ts *TestService) getScalingTimeout() time.Duration {
	scalingTimeout := time.Duration(math.Max(float64(ts.GetScale())*20, 120))
	return time.Second * scalingTimeout
}

// Get the current weight of a revision. If it does not exist, then it is 0.
func getRevisionWeight(app riov1.App, version string) int {
	if val, ok := app.Status.RevisionWeight[version]; ok {
		return val.Weight
	}
	return 0
}

//////////////////
// Wait helpers
//////////////////

// Wait until a service hits ready state, or error out
func (ts *TestService) waitForReadyService() error {
	f := wait.ConditionFunc(func() (bool, error) {
		err := ts.reload()
		if err == nil {
			if ts.GetAvailableReplicas() > 0 {
				return true, nil
			}
		}
		return false, nil
	})
	err := wait.Poll(2*time.Second, ts.getScalingTimeout(), f)
	if err != nil {
		return fmt.Errorf("service %v never reached ready status", ts.Name)
	}
	ts.reload()
	return nil
}

// Wait until a service has finished building, or error out
func (ts *TestService) waitForBuild() error {
	f := wait.ConditionFunc(func() (bool, error) {
		err := ts.reloadBuild()
		if err == nil {
			if ts.Build.Status.GetCondition(apis.ConditionSucceeded) != nil && ts.Build.Status.GetCondition(apis.ConditionSucceeded).IsTrue() {
				return true, nil
			}
		}
		return false, nil
	})
	err := wait.Poll(10*time.Second, 240*time.Second, f)
	if err != nil {
		return fmt.Errorf("build never completed for service: %v. Error: %v", ts.Name, err.Error())
	}
	return nil
}

// Wait until a service hits wanted number of available replicas, or error out
func (ts *TestService) waitForAvailableReplicas(want int) error {
	f := wait.ConditionFunc(func() (bool, error) {
		err := ts.reload()
		if err == nil {
			if ts.GetAvailableReplicas() == want {
				return true, nil
			}
		}
		return false, nil
	})
	err := wait.Poll(2*time.Second, ts.getScalingTimeout(), f)
	if err != nil {
		return fmt.Errorf("service %v failed to scale up available replicas", ts.Name)
	}
	return nil
}

// Wait until a service's scale field hits wanted int, or error out
func (ts *TestService) waitForScale(want int) error {
	f := wait.ConditionFunc(func() (bool, error) {
		err := ts.reload()
		if err == nil {
			if ts.GetScale() == want {
				return true, nil
			}
		}
		return false, nil
	})
	err := wait.Poll(2*time.Second, 60*time.Second, f)
	if err != nil {
		return errors.New("service failed to scale")
	}
	return nil
}

// Wait until a service has actual weight we want, note this is not spec weight which changes immediately
func (ts *TestService) waitForWeight(percentage int) error {
	f := wait.ConditionFunc(func() (bool, error) {
		ts.reloadApp()
		if val, ok := ts.App.Status.RevisionWeight[ts.Version]; ok {
			if val.Weight == percentage {
				return true, nil
			}
		}
		return false, nil
	})
	err := wait.Poll(2*time.Second, 60*time.Second, f)
	if err != nil {
		return errors.New("service revision never reached goal weight")
	}
	return nil
}

func (ts *TestService) waitForEndpointDNS() (string, error) {
	if len(ts.Service.Status.Endpoints) > 0 {
		return ts.Service.Status.Endpoints[0], nil
	}
	f := wait.ConditionFunc(func() (bool, error) {
		err := ts.reload()
		if err == nil {
			if len(ts.Service.Status.Endpoints) > 0 {
				return true, nil
			}
		}
		return false, nil
	})
	err := wait.Poll(2*time.Second, 60*time.Second, f)
	if err != nil {
		return "", errors.New("service endpoint never created")
	}
	return ts.Service.Status.Endpoints[0], nil
}

func (ts *TestService) waitForAppEndpointDNS() (string, error) {
	if len(ts.App.Status.Endpoints) > 0 {
		return ts.App.Status.Endpoints[0], nil
	}
	f := wait.ConditionFunc(func() (bool, error) {
		err := ts.reloadApp()
		if err == nil {
			if len(ts.App.Status.Endpoints) > 0 {
				return true, nil
			}
		}
		return false, nil
	})
	err := wait.Poll(2*time.Second, 60*time.Second, f)
	if err != nil {
		return "", errors.New("app endpoint never created")
	}
	return ts.App.Status.Endpoints[0], nil
}
