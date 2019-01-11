package ns

import (
	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/pkg/namespace"
	"github.com/rancher/rio/pkg/stacknamespace"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	corev1 "github.com/rancher/types/apis/core/v1"
	"k8s.io/api/core/v1"
)

const (
	projectID = "field.cattle.io/projectId"
)

func Populate(currentNs *v1.Namespace, stack *riov1.Stack, output *objectset.ObjectSet) {
	ns := corev1.NewNamespace("", namespace.StackToNamespace(stack), v1.Namespace{})

	if project, ok := currentNs.Annotations[projectID]; ok {
		ns.Annotations = map[string]string{
			projectID: project,
		}
	}

	stacknamespace.SetStackLabels(stack, ns)
	output.Add(ns)
}
