package integration

import (
	"testing"

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
	adminUser.Create()
	privilegedUser.Create()
	standardUser.Create()
	readonlyUser.Create()

	it.Before(func() {})

	it.After(func() {
		testService.Remove()
	})

	when("user tries to create services with specific roles like rio-admin,rio-privileged,rio-standard,rio-readonly", func() {
		// TODO: Create a test to distinguish between admin and privileged
		it("rio-admin user should be to create disabled service-mesh and privileged services", func() {
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
			if err == nil {
				t.Fatal("rio-standard should not be able to create service that enable service mesh")
			}
			assert.Contains(t, err.Error(), insuffienctPrivilegesMsg)
			assert.Contains(t, err.Error(), "rio-servicemesh")
		})

		it("rio-standard should not be able to create host-network services", func() {
			testService.Kubeconfig = standardUser.Kubeconfig
			err := testService.CreateExpectingError(t, "--net", "host", "nginx")
			if err == nil {
				t.Fatal("rio-standard should not be able to create service that enable host networking")
			}
			assert.Contains(t, err.Error(), insuffienctPrivilegesMsg)
			assert.Contains(t, err.Error(), "rio-hostnetwork")
		})

		it("rio-readonly should not be able to create services", func() {
			testService.Kubeconfig = readonlyUser.Kubeconfig
			err := testService.CreateExpectingError(t, "nginx")
			if err == nil {
				t.Fatal("rio-readonly user should not be able to create services")
			}
			assert.Contains(t, err.Error(), "is forbidden")
		})

		it("rio-standard user should not be able to escalate privilege on global permissions", func() {
			testService.Kubeconfig = standardUser.Kubeconfig
			err := testService.CreateExpectingError(t, "--global-permission", "list rio.cattle.io/services", "nginx")
			if err == nil {
				t.Fatal("rio-standard should not be able to escalate privilege on global permissions")
			}
			assert.Contains(t, err.Error(), insuffienctPrivilegesMsg)
		})

		it("rio-standard user should not be able to create privileges it doesn't have in the current namespace", func() {
			testService.Kubeconfig = standardUser.Kubeconfig
			err := testService.CreateExpectingError(t, "--permission", "update admin.rio.cattle.io/publicdomain", "--no-mesh", "nginx")
			if err == nil {
				t.Fatal("rio-standard should not be able to create privileges it doesn't have")
			}
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

	}, spec.Flat())
}
