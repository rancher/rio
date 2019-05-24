package populate

import (
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/name"
	"github.com/rancher/wrangler/pkg/objectset"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Ingress(systemNamespace string, publicdomain *adminv1.PublicDomain, os *objectset.ObjectSet) {
	os.Add(constructors.NewIngress(systemNamespace, name.SafeConcatName(publicdomain.Name, name.Hex(publicdomain.Spec.DomainName, 5)), v1beta1.Ingress{
		Spec: v1beta1.IngressSpec{
			Backend: &v1beta1.IngressBackend{
				ServiceName: constants.IstioGateway,
				ServicePort: intstr.FromInt(80),
			},
		},
	}))
}
