package populate

import (
	"fmt"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ServiceForRouter(r *riov1.Router, gatewayNamespace, gatewayName string, os *objectset.ObjectSet) error {
	service := constructors.NewService(r.Namespace, r.Name,
		v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app":     r.Name,
					"version": "v0",
				},
			},
			Spec: v1.ServiceSpec{
				Type:         v1.ServiceTypeExternalName,
				ExternalName: fmt.Sprintf("%s.%s", gatewayName, gatewayNamespace),
			},
		})

	os.Add(service)
	return nil
}
