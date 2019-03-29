package populate

import (
	"github.com/rancher/rio/modules/stack/controllers/stackns/populate/ns"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
)

func Stack(namespace *v1.Namespace, stack *riov1.Stack, output *objectset.ObjectSet) error {
	ns.Populate(namespace, stack, output)
	return nil
}
