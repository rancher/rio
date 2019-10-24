package testutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

type TestRoute struct {
	Router riov1.Router
	Name   string
	Path   string
	T      *testing.T
}

// Executes "rio route add routename.testing-namespace/{routePath} to {service}"
// This does not take a domain param, that is setup by default.
// domain is optional, if empty it will generate a random domain.
// routePath is optional, if empty it will set only domain.
func (tr *TestRoute) Add(t *testing.T, domain string, routePath string, action string, target TestService) {
	tr.T = t
	if domain == "" {
		domain = RandomString(5)
	}
	tr.Path = routePath
	tr.Name = fmt.Sprintf("%s/%s", testingNamespace, domain)
	route := fmt.Sprintf("%s.%s%s", domain, testingNamespace, routePath)
	_, err := RioCmd([]string{"router", "add", route, action, target.Name})
	if err != nil {
		tr.T.Fatalf("route add command failed:  %v", err.Error())
	}
	err = tr.waitForRoute()
	if err != nil {
		tr.T.Fatalf(err.Error())
	}
}

// Takes name of existing router and returns it as a TestRoute
func GetRoute(t *testing.T, name string, routePath string) TestRoute {
	tr := TestRoute{
		Router: riov1.Router{},
		Name:   fmt.Sprintf("%s/%s", testingNamespace, name),
		Path:   routePath,
		T:      t,
	}
	err := tr.waitForRoute()
	if err != nil {
		tr.T.Fatalf(err.Error())
	}
	return tr
}

// Executes "rio rm" for this route. This will delete all routes under this domain.
func (tr *TestRoute) Remove() {
	if tr.Router.Name != "" {
		_, err := RioCmd([]string{"rm", tr.Name})
		if err != nil {
			tr.T.Logf("failed to delete route:  %v", err.Error())
		}
	}
}

// GetEndpointResponse performs an http.get against the route's full domain and path and
// returns response if status code is 200, otherwise it errors out
func (tr *TestRoute) GetEndpointResponse() string {
	if len(tr.Router.Status.Endpoints) == 0 {
		tr.T.Fatal("router has no endpoint")
	}
	response, err := WaitForURLResponse(fmt.Sprintf("%s%s", tr.Router.Status.Endpoints[0], tr.Path))
	if err != nil {
		tr.T.Fatal(err.Error())
	}
	return response
}

// GetKubeEndpointResponse performs an http.get against the route's full domain and all paths on it
// Returns responses if status code is 200 for all of them, otherwise it errors out
func (tr *TestRoute) GetKubeEndpointResponse() string {
	// Get Router object using kubectl
	args := []string{"get", "routers", tr.Router.GetName(), "-n", testingNamespace, "-o", "json"}
	resultString, err := KubectlCmd(args)
	if err != nil {
		tr.T.Fatalf("Failed to get rio.cattle.io.routers:  %v", err.Error())
	}
	var results riov1.Router
	err = json.Unmarshal([]byte(resultString), &results)
	if err != nil {
		tr.T.Fatalf("Failed to unmarshal results:  %v", err.Error())
	}

	// Validate there is a base endpoint
	if len(results.Status.Endpoints) == 0 {
		tr.T.Fatal("router has no endpoint")
	}
	response, err := WaitForURLResponse(fmt.Sprintf("%s%s", results.Status.Endpoints[0], tr.Path))
	if err != nil {
		tr.T.Fatal(err.Error())
	}

	return response
}

//////////////////
// Private methods
//////////////////

func (tr *TestRoute) reload() error {
	out, err := RioCmd([]string{"inspect", "--format", "json", tr.Name})
	if err != nil {
		return err
	}
	tr.Router = riov1.Router{}
	if err := json.Unmarshal([]byte(out), &tr.Router); err != nil {
		return err
	}
	return nil
}

func (tr *TestRoute) waitForRoute() error {
	f := wait.ConditionFunc(func() (bool, error) {
		err := tr.reload()
		if err == nil {
			if len(tr.Router.Status.Endpoints) > 0 {
				return true, nil
			}
		}
		return false, nil
	})
	err := wait.Poll(2*time.Second, 60*time.Second, f)
	if err != nil {
		return errors.New("router and router endpoint not successfully initiated")
	}
	return nil
}
