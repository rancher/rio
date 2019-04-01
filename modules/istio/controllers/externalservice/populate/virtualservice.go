package populate

import (
	"github.com/rancher/rio/modules/istio/controllers/service/populate"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/serviceset"
	"github.com/rancher/wrangler/pkg/objectset"
)

func VirtualServiceForExternalService(namespace string, es *riov1.ExternalService, serviceSet *serviceset.ServiceSet,
	svc *riov1.Service, os *objectset.ObjectSet) {

	dests := populate.DestsForService(svc.Name, serviceSet.Service.Namespace, serviceSet)
	serviceVS := populate.VirtualServiceFromSpec(namespace, svc.Name, svc.Namespace, svc, dests...)

	// override host match with external service
	serviceVS.Spec.Hosts = []string{}
	serviceVS.Name = es.Name
	serviceVS.Namespace = es.Namespace
	os.Add(serviceVS)
}
