package testutil

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/rancher/rio/modules/service/controllers/service/populate/rbac"
	"github.com/rancher/wrangler/pkg/kubeconfig"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/client-go/kubernetes"
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
	kubeconfig string
}

func (u *TestUser) GetKubeconfig() string {
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
	binding := rbac.NewBinding(testingNamespace, u.Username, nil)
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
	if _, err := client.RbacV1().RoleBindings(testingNamespace).Create(binding); err != nil && !errors.IsAlreadyExists(err) {
		u.T.Fatal(err)
	}

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

	u.kubeconfig = tmpfile.Name()
	return u.kubeconfig
}

func (u *TestUser) Cleanup() {
	os.RemoveAll(u.kubeconfig)
}
