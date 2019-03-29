package ns

import (
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/pkg/stacknamespace"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
)

const (
	projectID = "field.cattle.io/projectId"
)

func Populate(currentNs *v1.Namespace, stack *riov1.Stack, output *objectset.ObjectSet) {
	ns := constructors.NewNamespace(stack.Name, v1.Namespace{})

	if project, ok := currentNs.Annotations[projectID]; ok {
		ns.Annotations = map[string]string{
			projectID: project,
		}
	}

	stacknamespace.SetStackLabels(stack, ns)
	output.Add(ns)
}
