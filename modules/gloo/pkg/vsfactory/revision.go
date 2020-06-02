package vsfactory

import (
	"fmt"
	"time"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/services"
	solov1 "github.com/solo-io/gloo/projects/gateway/pkg/api/v1/kube/apis/gateway.solo.io/v1"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	istiov1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
)

func (f *VirtualServiceFactory) ForRevision(svc *riov1.Service) ([]*solov1.VirtualService, error) {
	app, version := services.AppAndVersion(svc)

	target, err := f.getTarget(svc, f.systemNamespace)
	if err != nil || !target.valid() {
		return nil, err
	}

	vs := newGlooVirtualService(target.Namespace, target.Name, target.Hosts, target)

	if svc.Spec.RequestTimeoutSeconds != nil {
		t := time.Duration(int64(*svc.Spec.RequestTimeoutSeconds)) * time.Second
		if vs.Spec.VirtualHost.Routes[0].Options == nil {
			vs.Spec.VirtualHost.Routes[0].Options = &gloov1.RouteOptions{}
		}
		vs.Spec.VirtualHost.Routes[0].Options.Timeout = &t
	}

	f.InjectACME(vs)

	result := []*solov1.VirtualService{
		vs,
	}

	tls, err := f.FindTLS(svc.Namespace, app, version, target.Hosts)
	if err != nil {
		return result, err
	}

	for hostname, tls := range tls {
		if _, err := f.secretCache.Get(f.systemNamespace, tls); err == nil {
			result = append(result, tlsCopy(hostname, f.systemNamespace, tls, vs))
		}
	}

	return result, nil
}

func (f *VirtualServiceFactory) ForIstioRevision(svc *riov1.Service) (*istiov1alpha3.VirtualService, error) {
	app, version := services.AppAndVersion(svc)

	target, err := f.getTarget(svc, f.systemNamespace)
	if err != nil || !target.valid() {
		return nil, err
	}

	vs := newIstioVirtualService(svc.Namespace, fmt.Sprintf("%s-%s", app, version), stripPorts(target.Hosts), target)
	f.InjectACMEIstio(vs)
	return vs, nil
}
