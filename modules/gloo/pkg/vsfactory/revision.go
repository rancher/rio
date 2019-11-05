package vsfactory

import (
	"time"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/services"
	solov1 "github.com/solo-io/gloo/projects/gateway/pkg/api/v1/kube/apis/gateway.solo.io/v1"
)

func (f *VirtualServiceFactory) ForRevision(svc *riov1.Service) ([]*solov1.VirtualService, error) {
	app, version := services.AppAndVersion(svc)

	target, err := getTarget(svc, f.systemNamespace)
	if err != nil || !target.valid() {
		return nil, err
	}

	vs := newVirtualService(target.Namespace, target.Name, target.Hosts, target)

	if svc.Spec.RequestTimeoutSeconds != nil {
		t := time.Duration(int64(*svc.Spec.RequestTimeoutSeconds)) * time.Second
		vs.Spec.VirtualHost.Routes[0].RoutePlugins.Timeout = &t
	}

	if err := f.InjectACME(vs); err != nil {
		return nil, err
	}

	result := []*solov1.VirtualService{
		vs,
	}

	tls, err := f.findTLS(svc.Namespace, app, version, target.Hosts)
	if err != nil {
		return result, err
	}

	for hostname, tls := range tls {
		result = append(result, tlsCopy(hostname, f.systemNamespace, tls, vs))
	}

	return result, nil
}
