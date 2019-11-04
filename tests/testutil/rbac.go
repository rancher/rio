package testutil

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/rancher/rio/modules/service/controllers/service/populate/rbac"
	"github.com/rancher/wrangler/pkg/kubeconfig"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	AdminUserBindingName  = "rio-admin"
	AdminUserGroupName    = "rio:admin"
	StandardBindingName   = "rio-standard"
	StandardGroupName     = "rio:standard"
	PrivilegedBindingName = "rio-privileged"
	PrivilegedGroupName   = "rio:privileged"
	ReadonlyBindingName   = "rio-readonly"
	ReadonlyGroupName     = "rio:readonly"
)

type TestUser struct {
	Username   string
	Group      string
	T          *testing.T
	Kubeconfig string
}

func (u *TestUser) Create() {
	loader := kubeconfig.GetInteractiveClientConfig(os.Getenv("KUBECONFIG"))
	rawConfig, err := loader.RawConfig()
	if err != nil {
		u.T.Fatal(err)
	}
	restConfig, err := loader.ClientConfig()
	if err != nil {
		u.T.Fatal(err)
	}

	client := kubernetes.NewForConfigOrDie(restConfig)
	groupName := strings.Replace(u.Username, "-", ":", -1)
	binding := rbac.NewBinding(TestingNamespace, u.Username, nil)
	binding.Subjects = []rbacv1.Subject{
		{
			Kind:     rbacv1.GroupKind,
			APIGroup: rbacv1.GroupName,
			Name:     groupName,
		},
	}
	binding.RoleRef = rbacv1.RoleRef{
		APIGroup: rbacv1.GroupName,
		Kind:     "ClusterRole",
		Name:     u.Username,
	}

	client.RbacV1().RoleBindings(TestingNamespace).Create(binding)

	for _, user := range rawConfig.AuthInfos {
		user.Impersonate = u.Username
		user.ImpersonateGroups = []string{u.Group}
	}
	tmpfile, err := ioutil.TempFile("", "kubeconfig-")
	if err != nil {
		u.T.Fatal(err)
	}
	if err := clientcmd.WriteToFile(rawConfig, tmpfile.Name()); err != nil {
		u.T.Fatal(err)
	}

	u.Kubeconfig = tmpfile.Name()
}

func (u *TestUser) Cleanup() {
	os.RemoveAll(u.Kubeconfig)
}
