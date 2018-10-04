package ns

import (
	"github.com/rancher/rio/pkg/deploy/stackdef/output"
	"github.com/rancher/rio/pkg/namespace"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	projectID = "field.cattle.io/projectId"
)

func Populate(currentNs *v1.Namespace, stack *v1beta1.Stack, output *output.Deployment) {
	ns := &v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        namespace.StackToNamespace(stack),
			Annotations: map[string]string{},
		},
	}

	if project, ok := currentNs.Annotations[projectID]; ok {
		ns.Annotations[projectID] = project
	}

	output.Namespaces[ns.Name] = ns
}
