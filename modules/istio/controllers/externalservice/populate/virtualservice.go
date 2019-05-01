package populate

import (
	"github.com/rancher/rio/modules/istio/controllers/service/populate"
	v1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/serviceset"
	"github.com/rancher/wrangler/pkg/objectset"
)

func VirtualServiceForExternalService(namespace string, es *riov1.ExternalService, serviceSet *serviceset.ServiceSet, clusterDomain *v1.ClusterDomain,
	svc *riov1.Service, os *objectset.ObjectSet) {

	dests := populate.DestsForService(svc.Namespace, svc.Name, serviceSet)
	serviceVS := populate.VirtualServiceFromSpec(false, namespace, svc.Name, svc.Namespace, clusterDomain, svc, dests...)

	// override host match with external service
	serviceVS.Spec.Hosts = []string{}
	serviceVS.Name = es.Name
	serviceVS.Namespace = es.Namespace
	os.Add(serviceVS)
}
