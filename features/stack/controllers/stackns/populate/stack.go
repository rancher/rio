package populate

import (
	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/features/stack/controllers/stackns/populate/ns"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"k8s.io/api/core/v1"
)

func Stack(namespace *v1.Namespace, stack *riov1.Stack, output *objectset.ObjectSet) error {
	ns.Populate(namespace, stack, output)
	return nil
}
