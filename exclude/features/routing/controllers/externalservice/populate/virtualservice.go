package populate

import (
	"github.com/rancher/rio/features/routing/controllers/service/populate"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/serviceset"
	"github.com/rancher/wrangler/pkg/objectset"
)

func VirtualServiceForExternalService(systemNamespace string, es *riov1.ExternalService, serviceSet *serviceset.ServiceSet, svc *riov1.Service, stack *riov1.Stack, os *objectset.ObjectSet) {
	stackName := serviceSet.Service.Annotations["objectset.rio.cattle.io/owner-name"]
	serviceVS := populate.VsFromSpec(systemNamespace, stack, svc.Name, svc.Namespace, svc, populate.DestsForService(svc.Name, stackName, serviceSet)...)
	// override host match with external service
	serviceVS.Spec.Hosts = []string{}
	serviceVS.Name = es.Name
	serviceVS.Namespace = es.Namespace
	os.Add(serviceVS)
}
