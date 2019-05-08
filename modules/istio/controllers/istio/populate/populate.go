package populate

import (
	"github.com/rancher/wrangler/pkg/objectset"
)

func Istio(systemNamespace string) *objectset.ObjectSet {
	output := objectset.NewObjectSet()
	populateGateway(systemNamespace, output)

	return output
}
