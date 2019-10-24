package integration

import (
	"strings"
	"testing"

	"github.com/rancher/rio/tests/testutil"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
)

const (
	insuffienctPrivilegesMsg = "is attempting to grant RBAC permissions not currently held"
)

func rbacTests(t *testing.T, when spec.G, it spec.S) {
	var adminUser, privilegedUser, standardUser, readonlyUser *testutil.TestUser

	var testService testutil.TestService

	it.Before(func() {
		adminUser = &testutil.TestUser{
			Username: testutil.AdminUserBindingName,
			Group:    testutil.AdminUserGroupName,
			T:        t,
		}

		privilegedUser = &testutil.TestUser{
			Username: testutil.PrivilegedBindingName,
			Group:    testutil.PrivilegedGroupName,
			T:        t,
		}

		standardUser = &testutil.TestUser{
			Username: testutil.StandardBindingName,
			Group:    testutil.StandardGroupName,
			T:        t,
		}

		readonlyUser = &testutil.TestUser{
			Username: testutil.ReadonlyBindingName,
			Group:    testutil.ReadonlyGroupName,
			T:        t,
		}
	})

	it.After(func() {
		testService.Remove()
	})

	when("user tries to create services with specific roles like rio-admin,rio-privileged,rio-standard,rio-readonly", func() {
		it("rio-admin user should be to create service-mesh services", func() {
			testService.Kubeconfig = adminUser.GetKubeconfig()
			testService.Create(t, "nginx")
		})

		it("rio-privileged user should be able to create service-emesh services", func() {
			testService.Kubeconfig = privilegedUser.GetKubeconfig()
			testService.Create(t, "nginx")
		})

		// TODO: this needs to be changed when service-mesh api changes
		it("rio-standard should not be able to create service-mesh services", func() {
			testService.Kubeconfig = standardUser.GetKubeconfig()
			err := testService.CreateExpectingError(t, "nginx")
			if err == nil {
				t.Fatal("rio-standard should not be able to create service that enable service mesh")
			}
			assert.True(t, strings.Contains(err.Error(), insuffienctPrivilegesMsg))
			assert.True(t, strings.Contains(err.Error(), "rio-servicemesh"))
		})

		// TODO: we need to add more test for custom verb: hostport, hostnet, hostmount, serviceMesh and privilege
		it("rio-standard should not be able to create host-network services", func() {
			testService.Kubeconfig = standardUser.GetKubeconfig()
			err := testService.CreateExpectingError(t, "--net", "host", "nginx")
			if err == nil {
				t.Fatal("rio-standard should not be able to create service that enable host networking")
			}
			assert.True(t, strings.Contains(err.Error(), insuffienctPrivilegesMsg))
			assert.True(t, strings.Contains(err.Error(), "rio-hostnetwork"))
		})

		it("rio-readonly should not be able to create services", func() {
			testService.Kubeconfig = readonlyUser.GetKubeconfig()
			err := testService.CreateExpectingError(t, "nginx")
			if err == nil {
				t.Fatal("rio-readonly user should not be able to create services")
			}
			assert.True(t, strings.Contains(err.Error(), "is forbidden"))
		})

		it("rio-standard user should not be able to escalate privilege on global permissions", func() {
			testService.Kubeconfig = standardUser.GetKubeconfig()
			err := testService.CreateExpectingError(t, "--global-permission", "list rio.cattle.io/services", "nginx")
			if err == nil {
				t.Fatal("rio-standard should not be able to escalate privilege on global permissions")
			}
			assert.True(t, strings.Contains(err.Error(), insuffienctPrivilegesMsg))
		})

		it("rio-standard user should not be able to create privileges it doesn't have in the current namespace", func() {
			testService.Kubeconfig = standardUser.GetKubeconfig()
			err := testService.CreateExpectingError(t, "--permission", "update admin.rio.cattle.io/publicdomain", "--disable-service-mesh", "nginx")
			if err == nil {
				t.Fatal("rio-standard should not be able to create privileges it doesn't has")
			}
			assert.True(t, strings.Contains(err.Error(), insuffienctPrivilegesMsg))
		})

		it("rio-standard user should be able to create privileges it already has 1", func() {
			testService.Kubeconfig = standardUser.GetKubeconfig()
			testService.Create(t, "--permission", "list rio.cattle.io/services", "--disable-service-mesh", "nginx")
		})

		it("rio-standard user should be able to create privileges it already has 2", func() {
			testService.Kubeconfig = standardUser.GetKubeconfig()
			testService.Create(t, "--permission", "watch rio.cattle.io/app", "--disable-service-mesh", "nginx")
		})

	}, spec.Parallel(), spec.Flat())
}
