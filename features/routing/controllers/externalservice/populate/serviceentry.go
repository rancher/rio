package populate

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/rancher/norman/pkg/objectset"
	v1alpha3client "github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"istio.io/api/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ServiceEntry(stack *v1.Stack, esvc *v1.ExternalService, os *objectset.ObjectSet) error {
	u, err := parseTargetUrl(esvc.Spec.Target)
	if err != nil {
		return err
	}

	scheme := u.Scheme
	if scheme == "" {
		scheme = "http"
	}

	port, _ := strconv.ParseUint(u.Port(), 10, 64)
	if port == 0 {
		if scheme == "http" {
			port = 80
		} else if scheme == "https" {
			port = 443
		}
	}

	se := v1alpha3client.NewServiceEntry(esvc.Namespace, esvc.Name, v1alpha3client.ServiceEntry{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
			Labels: map[string]string{
				"rio.cattle.io/external-service": esvc.Name,
				"rio.cattle.io/stack":            stack.Name,
				"rio.cattle.io/project":          stack.Namespace,
			},
		},
		Spec: v1alpha3client.ServiceEntrySpec{
			Hosts:      []string{esvc.Spec.Target},
			Location:   v1alpha3.ServiceEntry_MESH_EXTERNAL,
			Resolution: v1alpha3.ServiceEntry_DNS,
			Ports: []v1alpha3client.Port{
				{
					Protocol: "HTTP",
					Number:   80,
					Name:     "http",
				},
				{
					Protocol: "HTTPS",
					Number:   443,
					Name:     "https",
				},
			},
			Endpoints: []v1alpha3client.ServiceEntry_Endpoint{
				{
					Address: u.Host,
					Ports: map[string]uint32{
						scheme: uint32(port),
					},
				},
			},
		},
	})

	os.Add(se)
	return nil
}

func parseTargetUrl(target string) (*url.URL, error) {
	if !strings.HasPrefix(target, "https://") && !strings.HasPrefix(target, "http://") {
		target = "http://" + target
	}
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	return u, nil
}
