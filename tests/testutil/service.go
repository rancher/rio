package testutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
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
	Name       string
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
	_, err := RioCmdWithRetry(args, envs...)
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
	ts.App = ts.Name
	if len(source) == 0 {
		source = []string{"-p", "80/http", "nginx"}
	}
	portFound := false
	for _, s := range source {
		if s == "-p" || s == "--port" {
			portFound = true
		}
	}
	if !portFound {
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
func GetService(t *testing.T, name string, app string, version string) TestService {
	if version != "" {
		name = fmt.Sprintf("%s@%s", name, version)
	}
	ts := TestService{
		Service: riov1.Service{},
		Name:    name,
		App:     app,
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
	if ts.T == nil {
		return
	}
	if ts.Kubeconfig != "" {
		err := os.RemoveAll(ts.Kubeconfig)
		if err != nil {
			ts.T.Log(err.Error())
		}
	}
	if ts.Service.Status.DeploymentReady || ts.Service.Spec.Template {
		_, err := RioCmdWithRetry([]string{"rm", ts.Name})
		if err != nil {
			ts.T.Log(err.Error())
		}
	}
}

// Attach attaches to the service: `rio --namespace testing-ns attach <service name>` and appends each line of output to an array
func (ts *TestService) Attach() []string {
	results, err := RioCmdWithTail(15, []string{"attach", ts.Name})
	if err != nil {
		ts.T.Fatalf("Failed to get attach output: %v", err.Error())
	}
	return results
}

// BuildAndCreate builds a local image and runs a service using it
func (ts *TestService) BuildAndCreate(t *testing.T, imageName string, imageVersion string, args ...string) {
	err := ts.BuildImage(t, imageName, imageVersion, args...)
	if err != nil {
		ts.T.Fatal(err)
	}
	ts.Create(t, fmt.Sprintf("localhost:5442/%s/%s:%s", TestingNamespace, imageName, imageVersion))
}

// BuildAndStage builds a local image and stages it onto another running service
func (ts *TestService) BuildAndStage(t *testing.T, imageName string, imageVersion string, args ...string) TestService {
	err := ts.BuildImage(t, imageName, imageVersion, args...)
	if err != nil {
		ts.T.Fatal(err)
	}
	return ts.Stage(fmt.Sprintf("localhost:5442/%s/%s:%s", TestingNamespace, imageName, imageVersion), imageVersion)
}

// BuildAndCreate builds a local image and runs a service using it
func (ts *TestService) BuildImage(t *testing.T, imageName string, imageVersion string, args ...string) error {
	ts.T = t
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Could not retrieve working directory.  %v", err.Error())
	}
	if strings.Contains(pwd, "tests/integration") {
		pwd = "./fixtures/Dockerfile" //
	} else {
		pwd = "./tests/integration/fixtures/Dockerfile"
	}
	args = append([]string{"build", "-f", pwd, "-t", fmt.Sprintf("%s:%s", imageName, imageVersion)}, args...)
	_, err = RioCmd(args)
	if err != nil {
		return fmt.Errorf("Failed to build image %s:%s.  %v", imageName, imageVersion, err.Error())
	}
	return nil
}

// Call "rio scale ns/service={scaleTo}"
func (ts *TestService) Scale(scaleTo int) {
	_, err := RioCmdWithRetry([]string{"scale", fmt.Sprintf("%s=%d", ts.Name, scaleTo)})
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

// WeightWithoutWaiting calls "rio weight {args} service_name={weightSpec}" on this service.
func (ts *TestService) WeightWithoutWaiting(weightSpec int, args ...string) {
	args = append(
		[]string{"weight"},
		append(args, fmt.Sprintf("%s=%d", ts.Name, weightSpec))...)
	_, err := RioCmdWithRetry(args)
	if err != nil {
		ts.T.Fatalf("weight command failed:  %v", err.Error())
	}
}

// Weight calls "rio weight {args} service_name={weightSpec}" on this service and waits until the service weight reaches the desired value.
func (ts *TestService) Weight(weightSpec int, args ...string) {
	ts.WeightWithoutWaiting(weightSpec, args...)
	paused := false
	for _, a := range args {
		if a == "pause" || a == "pause=true" {
			paused = true
		}
	}
	if !paused {
		err := ts.waitForWeight(weightSpec)
		if err != nil {
			ts.T.Fatal(err.Error())
		}
		if weightSpec == 100 {
			time.Sleep(5 * time.Second)
		}
	}
}

// Call "rio stage --image={source} ns/name:{version}", this will return a new TestService
func (ts *TestService) Stage(source, version string) TestService {
	_, err := RioCmdWithRetry([]string{"stage", "--image", source, ts.App, version})
	if err != nil {
		ts.T.Fatalf("stage command failed:  %v", err.Error())
	}
	return ts.stageCheck(version)
}

// Same as stage but uses the colon style namespacing
func (ts *TestService) StageExec(source, version string) TestService {
	nsName := fmt.Sprintf("%s:%s", TestingNamespace, ts.App)
	_, err := RioExecuteWithRetry([]string{"stage", "--image", "ibuildthecloud/demo:v3", nsName, version})
	if err != nil {
		ts.T.Fatalf("stage command failed:  %v", err.Error())
	}
	return ts.stageCheck(version)
}

// Executes a faux stage with run: "rio run -n ng@v3 --weight 50 nginx"
func (ts *TestService) StageRun(source, version, port, weight string) TestService {
	stageName := fmt.Sprintf("%s@%s", ts.App, version)
	_, err := RioCmdWithRetry([]string{"run", "-n", stageName, "--weight", weight, "-p", port, source})
	if err != nil {
		ts.T.Fatalf("stage command failed:  %v", err.Error())
	}
	return ts.stageCheck(version)
}

func (ts *TestService) stageCheck(version string) TestService {
	stagedService := TestService{
		T:       ts.T,
		App:     ts.App,
		Name:    fmt.Sprintf("%s@%s", ts.Name, version),
		Version: version,
	}
	err := stagedService.waitForReadyService()
	if err != nil {
		ts.T.Fatalf(err.Error())
	}
	return stagedService
}

// Promote calls "rio promote [args] service_name" to instantly promote a revision
func (ts *TestService) Promote(args ...string) {
	args = append(
		[]string{"promote", "--pause=false", "--duration=0s"},
		append(args, ts.Name)...)
	_, err := RioCmdWithRetry(args)
	if err != nil {
		ts.T.Fatalf("promote command failed:  %v", err.Error())
	}
	err = ts.waitForWeight(100)
	if err != nil {
		ts.T.Fatalf(err.Error())
	}
	time.Sleep(5 * time.Second)
	err = ts.waitForAppMatchesService()
	if err != nil {
		ts.T.Fatalf(err.Error())
	}
}

// Logs calls "rio logs ns/service" on this service
func (ts *TestService) Logs(args ...string) []string {
	args = append(
		[]string{"logs"},
		append(args, ts.Name)...)
	results, err := RioCmdWithTail(10, args)
	if err != nil {
		ts.T.Fatalf("Failed to get logs output: %v", err.Error())
	}
	return results
}

// Exec calls "rio exec ns/service {command}" on this service
func (ts *TestService) Exec(command ...string) string {
	out, err := RioCmdWithRetry(append([]string{"exec", "-it", ts.Name}, command...))
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
	response, err := WaitForURLResponse(ts.GetEndpointURLs()[0])
	if err != nil {
		ts.T.Fatal(err.Error())
	}
	return response
}

func (ts *TestService) GetAppEndpointResponse() string {
	response, err := WaitForURLResponse(ts.GetAppEndpointURLs()[0])
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
	ts.reload()
	if ts.Service.Spec.Weight != nil {
		return *ts.Service.Spec.Weight
	}
	return 0
}

// Return service's computed (actual) weight, not the spec (end-goal) weight
func (ts *TestService) GetCurrentWeight() int {
	out, err := RioCmdWithRetry([]string{"ps", "--format", "{{.Obj | id}}::{{.Data.Weight}}"})
	if err != nil {
		ts.T.Fatal(err)
	}
	for _, line := range strings.Split(out, "\n") {
		name := strings.Split(line, "::")[0]
		if name == ts.Name {
			weight, err := strconv.Atoi(strings.Split(line, "::")[1])
			if err != nil {
				ts.T.Fatal(err)
			}
			return weight
		}
	}
	return 0
}

// Return service's computed weight value
func (ts *TestService) GetComputedWeight() int {
	ts.reload()
	if ts.Service.Status.ComputedWeight != nil {
		return *ts.Service.Status.ComputedWeight
	}
	return 0
}

func (ts *TestService) GetImage() string {
	return ts.Service.Spec.Image
}

// Return RolloutDuration in seconds
func (ts *TestService) GetRolloutDuration() float64 {
	if ts.Service.Spec.RolloutDuration != nil {
		return ts.Service.Spec.RolloutDuration.Duration.Seconds()
	}
	return 0
}

// IsReady gets whether the service is created successfully and able to be used or not
func (ts *TestService) IsReady() bool {
	ts.reload()
	if ts.Service.Spec.Template {
		return true
	}
	if ts.Service.Spec.Autoscale != nil {
		return ts.Service.Status.DeploymentReady && ts.GetAvailableReplicas() >= int(*ts.Service.Spec.Autoscale.MinReplicas)
	}
	return ts.Service.Status.DeploymentReady && ts.GetAvailableReplicas() > 0
}

// GetRunningPods returns the kubectl overview of all running pods for this service in an array
// Each value in the array is a string, separated by spaces, that will have the Pod's NAME  READY  STATUS  RESTARTS  AGE in that order.
func (ts *TestService) GetRunningPods() []string {
	ts.reload()
	args := append([]string{"get", "pods",
		"-n", TestingNamespace,
		"-l", fmt.Sprintf("app=%s", ts.App),
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

// GetPodsAndReplicas waits until the service has reached its desired replica count, then returns the running pods and available replicas
func (ts *TestService) GetPodsAndReplicas() ([]string, int) {
	err := ts.waitForReadyService()
	if err != nil {
		ts.T.Fatalf("Service never fully reached ready state.  %v", err.Error())
	}
	if ts.Service.Status.ComputedReplicas != nil {
		_ = ts.waitForAvailableReplicas(*ts.Service.Status.ComputedReplicas)
	}
	return ts.GetRunningPods(), ts.GetAvailableReplicas()
}

// GenerateLoad queries the endpoint multiple times in order to put load on the service.
// It will execute for up to 120 seconds until there are ready pods on the service or the the AvailableReplicas equal the MaxScale
func (ts *TestService) GenerateLoad(timeIncrements string, concurrency int) {
	dur, _ := time.ParseDuration(timeIncrements)
	f := wait.ConditionFunc(func() (bool, error) {
		maxReplicas := 0
		minReplicas := 0
		if ts.Service.Spec.Autoscale != nil {
			maxReplicas = int(*ts.Service.Spec.Autoscale.MaxReplicas)
			minReplicas = int(*ts.Service.Spec.Autoscale.MinReplicas)
		}
		if maxReplicas > 0 {
			HeyCmd("http://"+GetHostname(ts.GetEndpointURLs()...), timeIncrements, concurrency)
			ts.reload()
			if ts.GetAvailableReplicas() > minReplicas {
				return true, nil
			}

		} else {
			HeyCmd("http://"+GetHostname(ts.GetEndpointURLs()...), timeIncrements, concurrency)
		}
		return false, nil
	})
	wait.Poll(dur, 120*time.Second, f)
}

// GetEndpointURLs returns the URLs for this service
func (ts *TestService) GetEndpointURLs() []string {
	endpoints, err := ts.waitForEndpointDNS()
	if err != nil {
		ts.T.Fatalf("Failed to get the endpoint url:  %v", err.Error())
	}
	return endpoints
}

// GetAppEndpointURLs retrieves the service's app endpoint URLs
func (ts *TestService) GetAppEndpointURLs() []string {
	endpoints, err := ts.waitForAppEndpointDNS()
	if err != nil {
		ts.T.Fatalf("Failed to get the endpoint url:  %v", err.Error())
	}
	return endpoints
}

// GetKubeEndpointURLs returns the app revision endpoint URLs as an array
func (ts *TestService) GetKubeEndpointURLs() []string {
	_, err := ts.waitForEndpointDNS()
	if err != nil {
		ts.T.Fatalf("Failed waiting for DNS:  %v", err.Error())
	}
	args := []string{"get", "service.rio.cattle.io",
		"-n", TestingNamespace,
		ts.Service.Name,
		"-o", `jsonpath="{.status.endpoints[*]}"`}
	urls, err := KubectlCmd(args)
	if err != nil {
		ts.T.Fatalf("Failed to get endpoint urls:  %v", err.Error())
	}
	return strings.Split(urls[1:len(urls)-1], " ")
}

// GetKubeFirstClusterDomain returns first cluster domain
func (ts *TestService) GetKubeFirstClusterDomain() (string, string) {
	args := []string{"get", "clusterdomains",
		"-o", `jsonpath='{.items[0].metadata.name}{"\t"}{.items[0].spec.addresses[0].ip}'`}
	clusterDomain, err := KubectlCmd(args)
	if err != nil {
		ts.T.Fatalf("Failed to get the first Cluster Domain:  %v", err.Error())
	}
	result := strings.Split(strings.Trim(clusterDomain, "'"), "\t")
	return strings.Trim(result[0], "\""), strings.Trim(result[1], "\"")
}

// GetKubeAppEndpointURLs returns the endpoint URL of the service's app
// by using kubectl and returns it as string
func (ts *TestService) GetKubeAppEndpointURLs() []string {
	_, err := ts.waitForAppEndpointDNS()
	if err != nil {
		ts.T.Fatalf("Failed waiting for DNS:  %v", err.Error())
	}
	args := []string{"get", "service.rio.cattle.io",
		"-n", TestingNamespace,
		ts.Service.Name,
		"-o", `jsonpath="{.status.appEndpoints[*]}"`}
	urls, err := KubectlCmd(args)
	if err != nil {
		ts.T.Fatalf("Failed to get app endpoint urls:  %v", err.Error())
	}
	return strings.Split(urls[1:len(urls)-1], " ")
}

// PodsResponsesMatchAvailableReplicas does a GetURL in the App endpoint and stores the response in a slice
// the length of the resulting slice should represent the number of responsive pods in a service.
// Returns true if the number of replicas is equal to the length of the responses slice.
func (ts *TestService) PodsResponsesMatchAvailableReplicas(path string, numberOfReplicas int) bool {
	i := 0
	replicasTimesRequests := numberOfReplicas * 8
	responses := make([]string, 0)
	for i < replicasTimesRequests {
		response, err := WaitForURLResponse(ts.GetEndpointURLs()[0] + path)
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

// GetKubeAvailableReplicas get the app number of ready replicasets with a clientset
// and returns true if that value match the scale given
func (ts *TestService) GetKubeAvailableReplicas() int {
	clientset := GetKubeClient(ts.T)
	replicaSetList, err := clientset.AppsV1().
		ReplicaSets(TestingNamespace).
		List(metav1.ListOptions{LabelSelector: "app=" + ts.App})
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

// WaitForDomain waits until either 1 minute has passed or the given domain is an available endpoint on the service or app
func (ts *TestService) WaitForDomain(domain string) error {
	f := wait.ConditionFunc(func() (bool, error) {
		err := ts.reload()
		if err == nil {
			for _, ep := range append(ts.Service.Status.Endpoints, ts.Service.Status.AppEndpoints...) {
				if ep == "http://"+domain {
					return true, nil
				}
			}
		}
		return false, nil
	})
	err := wait.Poll(2*time.Second, 60*time.Second, f)
	if err != nil {
		return fmt.Errorf("domain %v never added to service endpoints", domain)
	}
	return nil
}

//////////////////
// Private methods
//////////////////

// reload calls inspect on the service and uses that to reload our object
func (ts *TestService) reload() error {
	out, err := RioCmdWithRetry([]string{"inspect", "--format", "json", ts.Name})
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
	out, err := KubectlCmd([]string{"get", "taskrun", "-n", TestingNamespace, "-l", "gitwatcher.rio.cattle.io/service=" + ts.Service.GetName(), "-o", "json"})
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
	out, err := RioCmdWithRetry(args)
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
			if ts.IsReady() {
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
			if ts.Service.Spec.Template {
				return true, nil
			}
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
	err := wait.Poll(2*time.Second, 120*time.Second, f)
	if err != nil {
		return errors.New("service failed to scale")
	}
	return nil
}

// Wait until a service has the computed weight we want, note this is not spec weight but actual weight
func (ts *TestService) waitForWeight(target int) error {
	f := wait.ConditionFunc(func() (bool, error) {
		if ts.GetCurrentWeight() == target {
			return true, nil
		}
		return false, nil
	})
	err := wait.Poll(2*time.Second, 120*time.Second, f)
	if err != nil {
		return fmt.Errorf("service revision never reached goal weight. Expected %v. Got %v", target, ts.GetCurrentWeight())
	}
	return nil
}

func (ts *TestService) waitForEndpointDNS() ([]string, error) {
	ts.reload()
	if len(ts.Service.Status.Endpoints) > 0 {
		return ts.Service.Status.Endpoints, nil
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
		return []string{}, errors.New("service endpoint never created")
	}
	return ts.Service.Status.Endpoints, nil
}

func (ts *TestService) waitForAppEndpointDNS() ([]string, error) {
	ts.reload()
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

func (ts *TestService) waitForAppMatchesService() error {
	ts.reload()
	f := wait.ConditionFunc(func() (bool, error) {
		err := ts.reload()
		if err == nil {
			if ts.GetAppEndpointResponse() == ts.GetEndpointResponse() {
				return true, nil
			}
		}
		return false, nil
	})
	err := wait.Poll(2*time.Second, 60*time.Second, f)
	if err != nil {
		return errors.New("app endpoint response did not match service endpoint response")
	}
	return nil
}
