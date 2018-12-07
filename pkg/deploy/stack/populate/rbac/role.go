package rbac

import (
	"github.com/rancher/rio/pkg/deploy/stack/input"
	"github.com/rancher/rio/pkg/deploy/stack/output"
	"github.com/rancher/rio/pkg/deploy/stack/populate/servicelabels"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Populate(stack *input.Stack, service *riov1.Service, output *output.Deployment) error {
	labels := servicelabels.RioOnlyServiceLabels(stack, service)
	addGlobalRoles(service.Name, stack.Namespace, labels, service.Spec.GlobalPermissions, output)
	addRoles(service.Name, stack.Namespace, labels, &service.Spec.ServiceUnversionedSpec, output)
	return nil
}

func ServiceAccountName(service *riov1.Service) string {
	if len(service.Spec.Permissions) == 0 && len(service.Spec.GlobalPermissions) == 0 {
		return ""
	}
	return service.Name
}

func addGlobalRoles(name, namespace string, labels map[string]string, globalPermissions []riov1.Permission, output *output.Deployment) {
	if len(globalPermissions) == 0 {
		return
	}

	role := newClusterRole(name, namespace, labels)
	for _, perm := range globalPermissions {
		if perm.Role != "" {
			binding := newGlobalBinding(name, namespace, labels)
			binding.Subjects = append(binding.Subjects, v1.Subject{
				Kind:      "ServiceAccount",
				Name:      name,
				Namespace: namespace,
			})
			binding.RoleRef = v1.RoleRef{
				Name:     perm.Role,
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
			}

			output.ClusterRoleBindings[binding.Name] = binding
			continue
		}
		rule := v1.PolicyRule{
			Verbs: perm.Verbs,
		}
		if perm.URL == "" {
			rule.Resources = []string{perm.Resource}
			rule.APIGroups = []string{perm.APIGroup}
			if perm.Name != "" {
				rule.ResourceNames = []string{perm.Name}
			}
		} else {
			rule.NonResourceURLs = []string{perm.URL}
		}

		role.Rules = append(role.Rules, rule)
	}

	if len(role.Rules) > 0 {
		output.ClusterRoles[role.Name] = role
	}
}

func addRoles(name, namespace string, labels map[string]string, service *riov1.ServiceUnversionedSpec, output *output.Deployment) {
	if len(service.Permissions) == 0 && len(service.GlobalPermissions) == 0 {
		return
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
		if i < len(service.Permissions) {
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

			output.RoleBindings[binding.Name] = binding
		}
	}

	if len(role.Rules) > 0 {
		output.Roles[role.Name] = role

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

		output.RoleBindings[binding.Name] = binding
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

		output.ClusterRoleBindings[binding.Name] = binding
	}

	output.ServiceAccounts[serviceAccount.Name] = serviceAccount
}

func newServiceAccount(name, namespace string, labels map[string]string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
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
