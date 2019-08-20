package rbac

import (
	"github.com/rancher/rio/modules/service/controllers/service/populate/servicelabels"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/name"
	"github.com/rancher/wrangler/pkg/objectset"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Populate(service *riov1.Service, os *objectset.ObjectSet) error {
	labels := servicelabels.ServiceLabels(service)
	subject := subject(service)
	if subject == nil {
		return nil
	}

	addServiceAccount(labels, *subject, os)
	addRoles(labels, *subject, service, os)
	addRules(labels, *subject, service, os)
	addClusterRules(labels, *subject, service, os)
	return nil
}

func subject(service *riov1.Service) *v1.Subject {
	name := ServiceAccountName(service)
	if name == "" {
		return nil
	}

	return &v1.Subject{
		Name:      name,
		Namespace: service.Namespace,
		Kind:      "ServiceAccount",
	}

}

func ServiceAccountName(service *riov1.Service) string {
	if len(service.Spec.Permissions) == 0 &&
		len(service.Spec.GlobalPermissions) == 0 {
		return ""
	}
	return service.Name
}

func addRoles(labels map[string]string, subject v1.Subject, service *riov1.Service, os *objectset.ObjectSet) {
	for _, role := range service.Spec.Permissions {
		if role.Role == "" {
			continue
		}
		roleBinding := NewBinding(service.Namespace, name.SafeConcatName("rio", service.Name, role.Role), labels)
		roleBinding.Subjects = []v1.Subject{
			subject,
		}
		roleBinding.RoleRef = v1.RoleRef{
			Name:     role.Role,
			Kind:     "Role",
			APIGroup: "rbac.authorization.k8s.io",
		}
		os.Add(roleBinding)
	}

	for _, role := range service.Spec.GlobalPermissions {
		if role.Role == "" {
			continue
		}
		roleBinding := NewClusterBinding(name.SafeConcatName("rio", service.Namespace, service.Name, role.Role), labels)
		roleBinding.Subjects = []v1.Subject{
			subject,
		}
		roleBinding.RoleRef = v1.RoleRef{
			Name:     role.Role,
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
		}
		os.Add(roleBinding)
	}
}

func addRules(labels map[string]string, subject v1.Subject, service *riov1.Service, os *objectset.ObjectSet) {
	role := NewRole(service.Namespace, name.SafeConcatName("rio", service.Name), labels)
	for _, perm := range service.Spec.Permissions {
		if perm.Role != "" {
			continue
		}
		policyRule, ok := permToPolicyRule(perm)
		if ok {
			role.Rules = append(role.Rules, policyRule)
		}
	}

	if len(role.Rules) > 0 {
		os.Add(role)

		roleBinding := NewBinding(service.Namespace, name.SafeConcatName("rio", service.Name, role.Name), labels)
		roleBinding.Subjects = []v1.Subject{
			subject,
		}
		roleBinding.RoleRef = v1.RoleRef{
			Name:     role.Name,
			Kind:     "Role",
			APIGroup: "rbac.authorization.k8s.io",
		}
		os.Add(roleBinding)
	}
}

func addClusterRules(labels map[string]string, subject v1.Subject, service *riov1.Service, os *objectset.ObjectSet) {
	role := NewClusterRole(name.SafeConcatName("rio", service.Namespace, service.Name), labels)
	for _, perm := range service.Spec.GlobalPermissions {
		if perm.Role != "" {
			continue
		}
		policyRule, ok := permToPolicyRule(perm)
		if ok {
			role.Rules = append(role.Rules, policyRule)
		}
	}

	if len(role.Rules) > 0 {
		os.Add(role)

		roleBinding := NewClusterBinding(name.SafeConcatName("rio", service.Namespace, service.Name, role.Name), labels)
		roleBinding.Subjects = []v1.Subject{
			subject,
		}
		roleBinding.RoleRef = v1.RoleRef{
			Name:     role.Name,
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
		}
		os.Add(roleBinding)
	}
}

func permToPolicyRule(perm riov1.Permission) (v1.PolicyRule, bool) {
	policyRule := v1.PolicyRule{}
	valid := false

	if perm.Role != "" {
		return policyRule, valid
	}

	policyRule.Verbs = perm.Verbs
	if perm.URL == "" {
		if perm.ResourceName != "" {
			valid = true
			policyRule.ResourceNames = []string{perm.ResourceName}
		}

		policyRule.APIGroups = []string{perm.APIGroup}

		if perm.Resource != "" {
			valid = true
			policyRule.Resources = []string{perm.Resource}
		}
	} else {
		valid = true
		policyRule.NonResourceURLs = []string{perm.URL}
	}

	return policyRule, valid
}

func addServiceAccount(labels map[string]string, subject v1.Subject, os *objectset.ObjectSet) {
	sa := newServiceAccount(subject.Namespace, subject.Name, labels)
	os.Add(sa)
}

func newServiceAccount(namespace, name string, labels map[string]string) *corev1.ServiceAccount {
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

func NewRole(namespace, name string, labels map[string]string) *v1.Role {
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

func NewClusterRole(name string, labels map[string]string) *v1.ClusterRole {
	return &v1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      labels,
			Annotations: map[string]string{},
		},
	}
}

func NewClusterBinding(name string, labels map[string]string) *v1.ClusterRoleBinding {
	return &v1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      labels,
			Annotations: map[string]string{},
		},
	}
}

func NewBinding(namespace, name string, labels map[string]string) *v1.RoleBinding {
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
