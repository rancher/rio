package webhook

import (
	"github.com/rancher/wrangler-api/pkg/generated/controllers/rbac"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type rbacRestGetter struct {
	rbac.Interface
}

func (r rbacRestGetter) GetRole(namespace, name string) (*rbacv1.Role, error) {
	return r.Interface.V1().Role().Get(namespace, name, metav1.GetOptions{})
}

func (r rbacRestGetter) ListRoleBindings(namespace string) ([]*rbacv1.RoleBinding, error) {
	rolebindings, err := r.Interface.V1().RoleBinding().List(namespace, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var rbs []*rbacv1.RoleBinding
	for i := range rolebindings.Items {
		rbs = append(rbs, &rolebindings.Items[i])
	}
	return rbs, nil
}

func (r rbacRestGetter) GetClusterRole(name string) (*rbacv1.ClusterRole, error) {
	return r.Interface.V1().ClusterRole().Get(name, metav1.GetOptions{})
}

func (r rbacRestGetter) ListClusterRoleBindings() ([]*rbacv1.ClusterRoleBinding, error) {
	clusterrolebindings, err := r.Interface.V1().ClusterRoleBinding().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var crbs []*rbacv1.ClusterRoleBinding
	for i := range clusterrolebindings.Items {
		crbs = append(crbs, &clusterrolebindings.Items[i])
	}
	return crbs, nil
}
