package istio

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/rancher/rio/pkg/deploy/stack/input"
	"github.com/rancher/rio/pkg/deploy/stack/output"
	"istio.io/api/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func serviceEntries(stack *input.Stack) ([]*output.IstioObject, error) {
	var seResults []*output.IstioObject
	// external services
	for _, esvc := range stack.ExternalServices {
		sc := newServiceEntries(stack, esvc.Name, stack.Namespace)
		spec := &v1alpha3.ServiceEntry{}
		spec.Hosts = []string{esvc.Spec.Target}
		spec.Location = v1alpha3.ServiceEntry_MESH_EXTERNAL
		u, err := parseTargetUrl(esvc.Spec.Target)
		if err != nil {
			return nil, err
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
		spec.Resolution = v1alpha3.ServiceEntry_DNS
		spec.Ports = []*v1alpha3.Port{
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
		}
		spec.Endpoints = append(spec.Endpoints, &v1alpha3.ServiceEntry_Endpoint{
			Address: u.Host,
			Ports: map[string]uint32{
				scheme: uint32(port),
			},
		})
		sc.Spec = spec
		seResults = append(seResults, sc)
	}
	return seResults, nil
}
func newServiceEntries(stack *input.Stack, name, namespace string) *output.IstioObject {
	return &output.IstioObject{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "ServiceEntry",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: map[string]string{},
			Labels: map[string]string{
				"rio.cattle.io/stack":     stack.Stack.Name,
				"rio.cattle.io/workspace": stack.Stack.Namespace,
			},
		},
	}
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
