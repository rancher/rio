package populate

import (
	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/features/routing/controllers/service/populate"
	"github.com/rancher/rio/pkg/serviceset"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

func VirtualServiceForExternalService(es *riov1.ExternalService, serviceSet *serviceset.ServiceSet, svc *riov1.Service, stack *riov1.Stack, os *objectset.ObjectSet) {
	stackName := serviceSet.Service.Annotations["objectset.rio.cattle.io/owner-name"]
	projectName := serviceSet.Service.Annotations["objectset.rio.cattle.io/owner-namespace"]
	serviceVS := populate.VsFromSpec(stack, svc.Name, svc.Namespace, svc, populate.DestsForService(svc.Name, stackName, projectName, serviceSet)...)
	// override host match with external service
	serviceVS.Spec.Hosts = []string{}
	serviceVS.Name = es.Name
	serviceVS.Namespace = es.Namespace
	os.Add(serviceVS)
}
