package populate

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/rio/modules/istio/pkg/parse"

	"github.com/knative/pkg/apis/istio/v1alpha3"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
	v1alpha32 "istio.io/api/networking/v1alpha3"
)

func ServiceEntry(svc *v1.ExternalService, os *objectset.ObjectSet) error {
	if svc.Spec.FQDN == "" {
		return nil
	}

	se, err := serviceEntryForFQDN(svc.Spec.FQDN, svc)
	if err != nil {
		return err
	}
	os.Add(se)

	return nil
}

func serviceEntryForFQDN(fqdn string, svc *v1.ExternalService) (*v1alpha3.ServiceEntry, error) {
	u, err := parse.TargetURL(fqdn)
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
			Hosts:      []string{u.Hostname()},
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
					Address: u.Hostname(),
					Ports: map[string]uint32{
						scheme: uint32(port),
					},
				},
			},
		},
	})
	return se, nil
}
