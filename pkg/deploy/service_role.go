package deploy

import (
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func addGlobalRoles(objects []runtime.Object, name, namespace string, labels map[string]string, service *v1beta1.ServiceUnversionedSpec) ([]runtime.Object, error) {
	if len(service.GlobalPermissions) == 0 {
		return objects, nil
	}

	role := newClusterRole(name, namespace, labels)
	for _, perm := range service.GlobalPermissions {
		if perm.Role != "" {
			continue
		}
		rule := v1.PolicyRule{
			Verbs:     perm.Verbs,
			Resources: []string{perm.Resource},
			APIGroups: []string{perm.APIGroup},
		}
		if perm.Name != "" {
			rule.ResourceNames = []string{perm.Name}
		}

		role.Rules = append(role.Rules, rule)
	}

	if len(role.Rules) > 0 {
		objects = append(objects, role)
	}

	return objects, nil
}

func addRoles(objects []runtime.Object, name, namespace string, labels map[string]string, service *v1beta1.ServiceUnversionedSpec) ([]runtime.Object, bool, error) {
	if len(service.Permissions) == 0 && len(service.GlobalPermissions) == 0 {
		return objects, false, nil
	}

	t := true
	serviceAccount := newServiceAccount(name, namespace, labels)
	serviceAccount.AutomountServiceAccountToken = &t

	role := newRole(name, namespace, labels)
	for _, perm := range service.Permissions {
		if perm.Role != "" {
			continue
		}
		rule := v1.PolicyRule{
			Verbs:     perm.Verbs,
			Resources: []string{perm.Resource},
			APIGroups: []string{perm.APIGroup},
		}
		if perm.Name != "" {
			rule.ResourceNames = []string{perm.Name}
		}

		role.Rules = append(role.Rules, rule)
	}

	needsGlobalRoleBinding := false
	for i, perm := range append(service.Permissions, service.GlobalPermissions...) {
		if perm.Role == "" {
			if i >= len(service.Permissions) {
				needsGlobalRoleBinding = true
			}
			continue
		}

		binding := newBinding(name+"-"+perm.Role, namespace, labels)
		binding.Subjects = append(binding.Subjects, v1.Subject{
			Kind:      serviceAccount.Kind,
			Name:      name,
			Namespace: namespace,
		})
		binding.RoleRef = v1.RoleRef{
			Name:     perm.Role,
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     role.Kind,
		}

		objects = append(objects, binding)
	}

	if len(role.Rules) > 0 {
		objects = append(objects, role)

		binding := newBinding(name, namespace, labels)
		binding.Subjects = append(binding.Subjects, v1.Subject{
			Kind:      serviceAccount.Kind,
			Name:      name,
			Namespace: namespace,
		})
		binding.RoleRef = v1.RoleRef{
			Name:     name,
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     role.Kind,
		}

		objects = append(objects, binding)
	}

	if needsGlobalRoleBinding {
		binding := newGlobalBinding(name, namespace, labels)
		binding.Subjects = append(binding.Subjects, v1.Subject{
			Kind:      serviceAccount.Kind,
			Name:      name,
			Namespace: namespace,
		})
		binding.RoleRef = v1.RoleRef{
			Name:     name + "-" + namespace,
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
		}

		objects = append(objects, binding)
	}

	objects = append(objects, serviceAccount)
	return objects, true, nil
}

func newServiceAccount(name, namespace string, labels map[string]string) *v12.ServiceAccount {
	return &v12.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: map[string]string{},
		},
	}
}

func newRole(name, namespace string, labels map[string]string) *v1.Role {
	return &v1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: map[string]string{},
		},
	}
}

func newClusterRole(name, namespace string, labels map[string]string) *v1.ClusterRole {
	return &v1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name + "-" + namespace,
			Labels:      labels,
			Annotations: map[string]string{},
		},
	}
}

func newGlobalBinding(name, namespace string, labels map[string]string) *v1.ClusterRoleBinding {
	return &v1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name + "-" + namespace,
			Labels:      labels,
			Annotations: map[string]string{},
		},
	}
}

func newBinding(name, namespace string, labels map[string]string) *v1.RoleBinding {
	return &v1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: map[string]string{},
		},
	}
}
