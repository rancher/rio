package integration

import (
	"testing"
	"time"

	"github.com/rancher/rio/tests/testutil"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
)

const (
	insuffienctPrivilegesMsg = "is attempting to grant RBAC permissions not currently held"
)

func rbacTests(t *testing.T, when spec.G, it spec.S) {
	adminUser := &testutil.TestUser{
		Username: testutil.AdminUserBindingName,
		Group:    testutil.AdminUserGroupName,
		T:        t,
	}

	privilegedUser := &testutil.TestUser{
		Username: testutil.PrivilegedBindingName,
		Group:    testutil.PrivilegedGroupName,
		T:        t,
	}

	standardUser := &testutil.TestUser{
		Username: testutil.StandardBindingName,
		Group:    testutil.StandardGroupName,
		T:        t,
	}

	readonlyUser := &testutil.TestUser{
		Username: testutil.ReadonlyBindingName,
		Group:    testutil.ReadonlyGroupName,
		T:        t,
	}
	var testService testutil.TestService
	var riofile testutil.TestRiofile
	adminUser.Create()
	privilegedUser.Create()
	standardUser.Create()
	readonlyUser.Create()
	// sleep two seconds to make sure binding takes effect
	time.Sleep(time.Second * 5)

	it.Before(func() {})

	it.After(func() {
		riofile.Remove()
		testService.Remove()
	})

	when("user tries to create services/stacks with specific roles like rio-admin,rio-privileged,rio-standard,rio-readonly", func() {
		// TODO: Create a test to distinguish between admin and privileged
		it("rio-admin user should be to create service-mesh services", func() {
			testService.Kubeconfig = adminUser.Kubeconfig
			testService.Create(t, "--no-mesh", "--privileged", "nginx")
			assert.True(t, testService.IsReady())
		})

		it("rio-privileged user should be able to create disabled service-mesh and privileged services", func() {
			testService.Kubeconfig = privilegedUser.Kubeconfig
			testService.Create(t, "--no-mesh", "--privileged", "nginx")
			assert.True(t, testService.IsReady())
		})

		it("rio-standard should not be able to create disabled service-mesh services", func() {
			var testService testutil.TestService
			testService.Kubeconfig = standardUser.Kubeconfig
			err := testService.CreateExpectingError(t, "--no-mesh", "nginx")
			assert.Error(t, err, "rio-standard should not be able to create service that enable service mesh")
			assert.Contains(t, err.Error(), insuffienctPrivilegesMsg)
			assert.Contains(t, err.Error(), "rio-servicemesh")
		})

		it("rio-standard should not be able to create host-network services", func() {
			testService.Kubeconfig = standardUser.Kubeconfig
			err := testService.CreateExpectingError(t, "--net", "host", "nginx")
			assert.Error(t, err, "rio-standard should not be able to create service that enable host networking")
			assert.Contains(t, err.Error(), insuffienctPrivilegesMsg)
			assert.Contains(t, err.Error(), "rio-hostnetwork")
		})

		it("rio-standard should not be able to create host-port services", func() {
			testService.Kubeconfig = standardUser.Kubeconfig
			err := testService.CreateExpectingError(t, "-p", "80,hostport", "nginx")
			assert.Error(t, err, "rio-standard should not be able to create service that enable hostport")
			assert.Contains(t, err.Error(), insuffienctPrivilegesMsg)
			assert.Contains(t, err.Error(), "rio-hostport")
		})

		it("rio-standard should not be able to create privileged services", func() {
			testService.Kubeconfig = standardUser.Kubeconfig
			err := testService.CreateExpectingError(t, "--privileged", "nginx")
			assert.Error(t, err, "rio-standard should not be able to create service that enable privileged")
			assert.Contains(t, err.Error(), insuffienctPrivilegesMsg)
			assert.Contains(t, err.Error(), "rio-privileged")
		})

		it("rio-standard should not be able to create hostpath services", func() {
			testService.Kubeconfig = standardUser.Kubeconfig
			err := testService.CreateExpectingError(t, "-v", "/foo:/bar,hostPathType=directoryorcreate", "nginx")
			assert.Error(t, err, "rio-standard should not be able to create service that enable host path")
			assert.Contains(t, err.Error(), insuffienctPrivilegesMsg)
			assert.Contains(t, err.Error(), "rio-hostpath")
		})

		it("rio-readonly should not be able to create services", func() {
			testService.Kubeconfig = readonlyUser.Kubeconfig
			err := testService.CreateExpectingError(t, "nginx")
			assert.Error(t, err, "rio-readonly user should not be able to create services")
			assert.Contains(t, err.Error(), "is forbidden")
		})

		it("rio-standard user should not be able to escalate privilege on global permissions", func() {
			testService.Kubeconfig = standardUser.Kubeconfig
			err := testService.CreateExpectingError(t, "--global-permission", "list rio.cattle.io/services", "nginx")
			assert.Error(t, err, "rio-standard should not be able to escalate privilege on global permissions")
			assert.Contains(t, err.Error(), insuffienctPrivilegesMsg)
		})

		it("rio-standard user should not be able to create privileges with permissions it doesn't have in the current namespace", func() {
			testService.Kubeconfig = standardUser.Kubeconfig
			err := testService.CreateExpectingError(t, "--global-permission", "update admin.rio.cattle.io/publicdomain", "--no-mesh", "nginx")
			assert.Error(t, err, "rio-standard should not be able to create privileges it doesn't have")
			assert.Error(t, err, "rio-standard should not be able to create privileges it doesn't have")
			assert.Contains(t, err.Error(), insuffienctPrivilegesMsg)
		})

		it("rio-standard user should be able to create privileges it already has 1", func() {
			testService.Kubeconfig = standardUser.Kubeconfig
			testService.Create(t, "--permission", "list rio.cattle.io/services", "nginx")
			assert.True(t, testService.IsReady())
		})

		it("rio-standard user should be able to create privileges it already has 2", func() {
			testService.Kubeconfig = standardUser.Kubeconfig
			testService.Create(t, "--permission", "watch rio.cattle.io/services", "nginx")
			assert.True(t, testService.IsReady())
		})

		it("rio-admin user should be to create stacks with permissions", func() {
			riofile.Kubeconfig = adminUser.Kubeconfig
			err := riofile.UpWithRepo(t, "https://github.com/rancher/rio-demo", "", "--permission", "update apps/deployments")
			assert.NoError(t, err)
		})

		it("rio-standard user should not be able to create stacks with permissions it doesn't have in the current namespace", func() {
			riofile.Kubeconfig = standardUser.Kubeconfig
			err := riofile.UpWithRepo(t, "https://github.com/rancher/rio-demo", "", "--permission", "update apps/deployments")
			assert.Error(t, err, "rio-standard should not be able to create privileges it doesn't have")
			assert.Contains(t, err.Error(), insuffienctPrivilegesMsg)
		})
	}, spec.Parallel(), spec.Flat())
}
