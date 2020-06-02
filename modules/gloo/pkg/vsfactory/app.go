package vsfactory

import (
	"sort"
	"time"

	"strings"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	solov1 "github.com/solo-io/gloo/projects/gateway/pkg/api/v1/kube/apis/gateway.solo.io/v1"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	istiov1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
)

func (f *VirtualServiceFactory) ForApp(namespace, appName string, svcs []*riov1.Service) ([]*solov1.VirtualService, error) {
	hostnames, targets, err := f.getTargetsForApp(svcs, f.systemNamespace)
	if err != nil {
		return nil, err
	}

	if len(hostnames) == 0 {
		return nil, nil
	}

	vs := newGlooVirtualService(namespace, appName, hostnames, targets...)
	sort.Slice(svcs, func(i, j int) bool {
		if svcs[i].Spec.Weight == nil || svcs[j].Spec.Weight == nil {
			return false
		}
		return *svcs[i].Spec.Weight < *svcs[j].Spec.Weight
	})

	if svcs[len(svcs)-1].Spec.RequestTimeoutSeconds != nil {
		if vs.Spec.VirtualHost.Routes[0].Options == nil {
			vs.Spec.VirtualHost.Routes[0].Options = &gloov1.RouteOptions{}
		}
		t := time.Duration(int64(*svcs[len(svcs)-1].Spec.RequestTimeoutSeconds)) * time.Second
		vs.Spec.VirtualHost.Routes[0].Options.Timeout = &t
	}

	f.InjectACME(vs)

	result := []*solov1.VirtualService{
		vs,
	}

	tls, err := f.FindTLS(namespace, appName, "", hostnames)
	if err != nil {
		return nil, err
	}

	for hostname, tls := range tls {
		if _, err := f.secretCache.Get(f.systemNamespace, tls); err == nil {
			result = append(result, tlsCopy(hostname, f.systemNamespace, tls, vs))
		}
	}

	return result, nil
}

func (f *VirtualServiceFactory) ForAppIstio(namespace, appName string, svcs []*riov1.Service) (*istiov1alpha3.VirtualService, *istiov1alpha3.DestinationRule, error) {
	hostnames, targets, err := f.getTargetsForApp(svcs, f.systemNamespace)
	if err != nil {
		return nil, nil, err
	}

	if len(hostnames) == 0 {
		return nil, nil, nil
	}

	vs := newIstioVirtualService(namespace, appName, stripPorts(hostnames), targets...)
	f.InjectACMEIstio(vs)
	dest := newIstioDestinationRule(namespace, appName, targets...)
	return vs, dest, nil
}

func stripPorts(hostnames []string) (result []string) {
	seen := map[string]bool{}
	for _, hostname := range hostnames {
		host := strings.SplitN(hostname, ":", 2)[0]
		if !seen[host] {
			seen[host] = true
			result = append(result, host)
		}
	}
	return result
}
