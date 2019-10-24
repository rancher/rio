package testutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	apis "knative.dev/pkg/apis"
)

type TestService struct {
	Name       string // namespace/name
	App        string
	Service    riov1.Service
	Build      tektonv1alpha1.TaskRun
	Version    string
	T          *testing.T
	Kubeconfig string
}

// Create generates a new rio service, named randomly in the testing namespace, and
// returns a new TestService with it attached. Guarantees ready state but not live endpoint
func (ts *TestService) Create(t *testing.T, source ...string) {
	args, envs := ts.createArgs(t, source...)
	_, err := RioCmd(args, envs...)
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
}

func (ts *TestService) CreateExpectingError(t *testing.T, source ...string) error {
	args, envs := ts.createArgs(t, source...)
	_, err := RioCmd(args, envs...)
	if err != nil {
		return err
	}
	return nil
}

func (ts *TestService) createArgs(t *testing.T, source ...string) ([]string, []string) {
	ts.T = t
	ts.Version = "v0"
	ts.Name = RandomString(5)
	if len(source) == 0 {
		source = []string{"nginx"}
	}
	// Ensure port 80/http unless the source is from github
	if !ts.isGithubSource(source...) {
		source = append([]string{"-p", "80/http"}, source...)
	}
	args := append([]string{"run", "-n", ts.Name}, source...)

	var envs []string
	if ts.Kubeconfig != "" {
		envs = []string{fmt.Sprintf("KUBECONFIG=%s", ts.Kubeconfig)}
	}

	return args, envs
}

func (ts *TestService) isGithubSource(source ...string) bool {
	return len(source) > 0 && source[len(source)-1][0:4] == "http" && strings.Contains(source[len(source)-1], "github")
}

// Takes name and version of existing service and returns loaded TestService
func GetService(t *testing.T, name string, version string) TestService {
	ts := TestService{
		Service: riov1.Service{},
		Name:    name,
		Version: version,
		T:       t,
	}
	err := ts.waitForReadyService()
	if err != nil {
		ts.T.Fatalf(err.Error())
	}
	return ts
}

// Remove calls "rio rm" on this service. Logs error but does not fail test.
func (ts *TestService) Remove() {
	if ts.Kubeconfig != "" {
		err := os.RemoveAll(ts.Kubeconfig)
		if err != nil {
			ts.T.Log(err.Error())
		}
	}
	if ts.Service.Status.DeploymentReady {
		_, err := RioCmd([]string{"rm", ts.Name})
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
	name := fmt.Sprintf("%s:%s", ts.Name, version)
	_, err := RioCmd([]string{"stage", "--image", source, name})
	if err != nil {
		ts.T.Fatalf("stage command failed:  %v", err.Error())
	}
	stagedService := TestService{
		T:       ts.T,
		Name:    fmt.Sprintf("%s-%s", ts.Name, version),
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
	args = append([]string{"logs"}, append(args, ts.Name, "-a")...)
	out, err := RioCmd(args)
	if err != nil {
		ts.T.Fatalf("logs command failed:  %v", err.Error())
	}
	return out
}

// Exec calls "rio exec ns/service {command}" on this service
func (ts *TestService) Exec(command ...string) string {
	out, err := RioCmd(append([]string{"exec", ts.Name}, command...))
	if err != nil {
		ts.T.Fatalf("exec command failed:  %v", err.Error())
	}
	return out
}

// Export calls "rio export {serviceName}" and returns that in a new TestService object
func (ts *TestService) Export() TestService {
	args := []string{"export", "--format", "json", ts.Name}
	service, err := ts.loadExport(args)
	if err != nil {
		ts.T.Fatal(err.Error())
	}
	return service
}

// ExportRaw works the same as export, but with --raw flag
func (ts *TestService) ExportRaw() TestService {
	args := []string{"export", "--raw", "--format", "json", ts.Name}
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
	response, err := WaitForURLResponse(ts.GetServiceEndpointURL()[0])
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
	if ts.Service.Status.DeploymentReady && ts.Service.Status.ScaleStatus != nil {
		return ts.Service.Status.ScaleStatus.Available
	}
	return 0
}

// Returns desired scale, different from current available replicas
func (ts *TestService) GetScale() int {
	if ts.Service.Spec.Replicas != nil {
		return *ts.Service.Spec.Replicas
	}
	return 0
}

// Return service's goal weight, this is different from weight service is currently at
func (ts *TestService) GetSpecWeight() int {
	if ts.Service.Spec.Weight != nil {
		return *ts.Service.Spec.Weight
	}
	return 0
}

// Return service's actual current weight, not the spec (end-goal) weight
func (ts *TestService) GetCurrentWeight() int {
	ts.reload()
	if ts.Service.Status.ComputedWeight != nil {
		return *ts.Service.Status.ComputedWeight
	}
	return 0
}

func (ts *TestService) GetImage() string {
	return ts.Service.Spec.Image
}

// GetRunningPods returns the kubectl overview of all running pods for this service in an array
// Each value in the array is a string, separated by spaces, that will have the Pod's NAME  READY  STATUS  RESTARTS  AGE in that order.
func (ts *TestService) GetRunningPods() []string {
	ts.reload()
	args := append([]string{"get", "pods",
		"-n", testingNamespace,
		"-l", fmt.Sprintf("rio.cattle.io/service==%s", ts.Service.Name),
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
// It will execute for up to 60 seconds until there are ready pods on the service or the the AvailableReplicas equal the MaxScale
func (ts *TestService) GenerateLoad() {
	f := wait.ConditionFunc(func() (bool, error) {
		maxReplicas := 0
		availablePods := 0
		if ts.Service.Spec.Autoscale != nil {
			maxReplicas = int(*ts.Service.Spec.Autoscale.MaxReplicas)
		}
		if ts.Service.Status.ScaleStatus != nil {
			availablePods = ts.Service.Status.ScaleStatus.Available
		}
		if maxReplicas > 0 {
			HeyCmd(GetHostname(ts.GetServiceEndpointURL()[0]), "5s", 10*maxReplicas)
			ts.reload()
			if availablePods > 0 || ts.GetAvailableReplicas() == maxReplicas {
				return true, nil
			}

		}
		return false, nil
	})
	wait.Poll(5*time.Second, 60*time.Second, f)
	availablePods := 0
	if ts.Service.Status.ScaleStatus != nil {
		availablePods = ts.Service.Status.ScaleStatus.Available
	}
	if availablePods > 0 {
		ts.waitForAvailableReplicas(availablePods)
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

// GetServiceEndpointURL retrieves the service's app endpoint URL and returns it as string
func (ts *TestService) GetServiceEndpointURL() []string {
	endpoints, err := ts.waitForAppEndpointDNS()
	if err != nil {
		ts.T.Fatalf("Failed to get the endpoint url:  %v", err.Error())
		return []string{}
	}
	return endpoints
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
	args := []string{"get", "apps", ts.Name, "-o", `jsonpath="{.status.endpoints[0]}"`}
	url, err := KubectlCmd(args)
	if err != nil {
		ts.T.Fatalf("Failed to get app endpoint url:  %v", err.Error())
		return ""
	}

	return strings.Replace(url, "\"", "", -1) // remove double quotes from output
}

// PodsResponsesMatchAvailableReplicas does a GetURL in the App endpoint and stores the response in a slice
// the length of the resulting slice should represent the number of responsive pods in a service.
// Returns true if the number of replicas is equal to the length of the responses slice.
func (ts *TestService) PodsResponsesMatchAvailableReplicas(path string, numberOfReplicas int) bool {
	i := 0
	replicasTimesRequests := numberOfReplicas * 8
	responses := make([]string, 0)
	for i < replicasTimesRequests {
		response, err := WaitForURLResponse(ts.GetServiceEndpointURL()[0] + path)
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
			if ts.Service.Spec.Autoscale != nil {
				if int(*ts.Service.Spec.Autoscale.MinReplicas) == ts.GetAvailableReplicas() {
					return true, nil
				}
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
	out, err := RioCmd([]string{"inspect", "--format", "json", ts.Name})
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

// Wait until a service has the computed weight we want, note this is not spec weight but actual weight
func (ts *TestService) waitForWeight(percentage int) error {
	f := wait.ConditionFunc(func() (bool, error) {
		ts.reload()
		if ts.Service.Status.ComputedWeight != nil {
			if percentage == *ts.Service.Status.ComputedWeight {
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

func (ts *TestService) waitForAppEndpointDNS() ([]string, error) {
	if len(ts.Service.Status.AppEndpoints) > 0 {
		return ts.Service.Status.AppEndpoints, nil
	}
	f := wait.ConditionFunc(func() (bool, error) {
		err := ts.reload()
		if err == nil {
			if len(ts.Service.Status.AppEndpoints) > 0 {
				return true, nil
			}
		}
		return false, nil
	})
	err := wait.Poll(2*time.Second, 60*time.Second, f)
	if err != nil {
		return nil, errors.New("app endpoint never created")
	}
	return ts.Service.Status.AppEndpoints, nil
}
