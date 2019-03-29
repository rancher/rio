package populate

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	v1alpha32 "istio.io/api/networking/v1alpha3"

	"github.com/rancher/rio/pkg/constructors"

	"github.com/knative/pkg/apis/istio/v1alpha3"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/objectset"
)

func ServiceEntry(svc *v1.ExternalService, stack *v1.Stack, os *objectset.ObjectSet) error {
	if svc.Spec.FQDN != "" {
		se, err := populateServiceEntryForFqdn(svc.Spec.FQDN, svc)
		if err != nil {
			return err
		}
		os.Add(se)
	}
	return nil
}

func populateServiceEntryForFqdn(fqdn string, svc *v1.ExternalService) (*v1alpha3.ServiceEntry, error) {
	u, err := ParseTargetUrl(fqdn)
	if err != nil {
		return nil, err
	}

	scheme := u.Scheme
	if scheme == "" {
		scheme = "http"
	}

	port, _ := strconv.Atoi(u.Port())
	if port == 0 {
		if scheme == "http" {
			port = 80
		} else if scheme == "https" {
			port = 443
		}
	}

	se := constructors.NewServiceEntry(svc.Namespace, svc.Name, v1alpha3.ServiceEntry{
		Spec: v1alpha3.ServiceEntrySpec{
			Hosts:      []string{u.Host},
			Location:   int32(v1alpha32.ServiceEntry_MESH_EXTERNAL),
			Resolution: int32(v1alpha32.ServiceEntry_DNS),
			Ports: []v1alpha3.Port{
				{
					Protocol: v1alpha3.PortProtocol(strings.ToUpper(scheme)),
					Number:   port,
					Name:     fmt.Sprintf("%s-%v", scheme, port),
				},
			},
			Endpoints: []v1alpha3.ServiceEntry_Endpoint{
				{
					Address: u.Host,
					Ports: map[string]uint32{
						scheme: uint32(port),
					},
				},
			},
		},
	})
	return se, nil
}

func ParseTargetUrl(target string) (*url.URL, error) {
	if !strings.HasPrefix(target, "https://") && !strings.HasPrefix(target, "http://") {
		target = "http://" + target
	}
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	return u, nil
}
