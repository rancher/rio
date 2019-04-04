package rbac

import (
	"github.com/rancher/rio/modules/service/controllers/service/populate/servicelabels"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/name"
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
		Kind:      "Subject",
		APIGroup:  "rbac.authorization.k8s.io",
	}

}

func ServiceAccountName(service *riov1.Service) string {
	if len(service.Spec.Roles) == 0 &&
		len(service.Spec.ClusterRoles) == 0 &&
		len(service.Spec.Rules) == 0 &&
		len(service.Spec.ClusterRules) == 0 {
		return ""
	}
	return service.Name
}

func addRoles(labels map[string]string, subject v1.Subject, service *riov1.Service, os *objectset.ObjectSet) {
	for _, role := range service.Spec.Roles {
		roleBinding := newBinding(service.Namespace, name.SafeConcatName("rio", service.Name, role), labels)
		roleBinding.Subjects = []v1.Subject{
			subject,
		}
		roleBinding.RoleRef = v1.RoleRef{
			Name:     role,
			Kind:     "Role",
			APIGroup: "rbac.authorization.k8s.io",
		}
		os.Add(roleBinding)
	}

	for _, role := range service.Spec.ClusterRoles {
		roleBinding := newClusterBinding(name.SafeConcatName("rio", service.Namespace, service.Name, role), labels)
		roleBinding.Subjects = []v1.Subject{
			subject,
		}
		roleBinding.RoleRef = v1.RoleRef{
			Name:     role,
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
		}
		os.Add(roleBinding)
	}
}

func addRules(labels map[string]string, subject v1.Subject, service *riov1.Service, os *objectset.ObjectSet) {
	if len(service.Spec.Roles) > 0 {
		role := newRole(service.Namespace, name.SafeConcatName("rio", service.Name), labels)
		role.Rules = service.Spec.Rules
		os.Add(role)

		roleBinding := newBinding(service.Namespace, name.SafeConcatName("rio", service.Name, role.Name), labels)
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

	if len(service.Spec.ClusterRoles) > 0 {
		role := newClusterRole(name.SafeConcatName("rio", service.Namespace, service.Name), labels)
		role.Rules = service.Spec.ClusterRules
		os.Add(role)

		roleBinding := newClusterBinding(name.SafeConcatName("rio", service.Namespace, service.Name, role.Name), labels)
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

func newRole(namespace, name string, labels map[string]string) *v1.Role {
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

func newClusterRole(name string, labels map[string]string) *v1.ClusterRole {
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

func newClusterBinding(name string, labels map[string]string) *v1.ClusterRoleBinding {
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

func newBinding(namespace, name string, labels map[string]string) *v1.RoleBinding {
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
