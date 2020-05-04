package vsfactory

import (
	"sort"
	"time"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	solov1 "github.com/solo-io/gloo/projects/gateway/pkg/api/v1/kube/apis/gateway.solo.io/v1"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
)

func (f *VirtualServiceFactory) ForApp(namespace, appName string, svcs []*riov1.Service) ([]*solov1.VirtualService, error) {
	hostnames, targets, err := f.getTargetsForApp(svcs, f.systemNamespace)
	if err != nil {
		return nil, err
	}

	if len(hostnames) == 0 {
		return nil, nil
	}

	vs := newVirtualService(namespace, appName, hostnames, targets...)
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

	if err := f.InjectACME(vs); err != nil {
		return nil, err
	}

	result := []*solov1.VirtualService{
		vs,
	}

	tls, err := f.FindTLS(namespace, appName, "", hostnames)
	if err != nil {
		return nil, err
	}

	for hostname, tls := range tls {
		result = append(result, tlsCopy(hostname, f.systemNamespace, tls, vs))
	}

	return result, nil
}
