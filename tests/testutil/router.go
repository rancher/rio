package testutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

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
// routePath is optional, if empty it will set only domain.
func (tr *TestRoute) Add(t *testing.T, routePath string, action string, target TestService) {
	tr.T = t
	fakeDomain := RandomString(5)
	tr.Path = routePath
	tr.Name = fmt.Sprintf("%s/%s", testingNamespace, fakeDomain)
	route := fmt.Sprintf("%s.%s%s", fakeDomain, testingNamespace, routePath)
	args := []string{"add", route, action, target.Name}
	_, err := RioCmd("route", args)
	if err != nil {
		tr.T.Fatalf("route add command failed:  %v", err.Error())
	}
	err = tr.waitForRoute()
	if err != nil {
		tr.T.Fatalf(err.Error())
	}
}

// Executes "rio rm" for this route. This will delete all routes under this domain.
func (tr *TestRoute) Remove() {
	if tr.Router.Name != "" {
		_, err := RioCmd("rm", []string{"--type", "router", tr.Name})
		if err != nil {
			tr.T.Logf("failed to delete route:  %v", err.Error())
		}
	}
}

// GetEndpoint performs an http.get against the route's full domain and path and
// returns response if status code is 200, otherwise it errors out
func (tr *TestRoute) GetEndpoint() string {
	if len(tr.Router.Status.Endpoints) == 0 {
		tr.T.Fatal("router has no endpoint")
	}
	response, err := WaitForURLResponse(fmt.Sprintf("%s%s", tr.Router.Status.Endpoints[0], tr.Path))
	if err != nil {
		tr.T.Fatal(err.Error())
	}
	return response
}

//////////////////
// Private methods
//////////////////

func (tr *TestRoute) reload() error {
	args := append([]string{"--type", "router", "--format", "json", tr.Name})
	out, err := RioCmd("inspect", args)
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
	f := func() bool {
		err := tr.reload()
		if err == nil {
			if len(tr.Router.Status.Endpoints) > 0 {
				return true
			}
		}
		return false
	}
	ok := WaitFor(f, 60)
	if ok == false {
		return errors.New("router and router endpoint not successfully initiated")
	}
	return nil
}
