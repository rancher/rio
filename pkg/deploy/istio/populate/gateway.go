package populate

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rancher/rio/pkg/certs"

	"github.com/rancher/rio/pkg/deploy/istio/input"
	istioOutput "github.com/rancher/rio/pkg/deploy/istio/output"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func populateGateway(input *input.IstioDeployment, output *istioOutput.Deployment) error {
	if !output.Enabled {
		return nil
	}

	if len(output.Ports) == 0 {
		return nil
	}

	gws := v1alpha3.GatewaySpec{
		Selector: settings.IstioGatewaySelector,
	}
	sslDir := GetSSLDir()

	filteredPorts := make([]string, 0)
	set := map[string]struct{}{}
	for _, port := range output.Ports {
		set[port] = struct{}{}
	}
	for p := range set {
		filteredPorts = append(filteredPorts, p)
	}
	http80Open := false
	for _, port := range filteredPorts {
		split := strings.SplitN(port, "/", 2)
		port, err := strconv.ParseInt(split[0], 10, 0)
		if err != nil {
			return err
		}
		protocol := "http"
		if len(split) == 2 {
			protocol = split[1]
		}

		if protocol == "http" {
			gws.Servers = append(gws.Servers, &v1alpha3.Server{
				Port: &v1alpha3.Port{
					Protocol: "HTTP",
					Number:   uint32(port),
					Name:     fmt.Sprintf("%v-%v", protocol, port),
				},
				Hosts: []string{"*"},
			})
			if port == 80 {
				http80Open = true
			}
		} else if protocol == "https" {
			gws.Servers = append(gws.Servers, &v1alpha3.Server{
				Port: &v1alpha3.Port{
					Protocol: "HTTPS",
					Number:   uint32(port),
					Name:     fmt.Sprintf("%v-%v", protocol, port),
				},
				Hosts: []string{fmt.Sprintf("*.%s", settings.ClusterDomain.Get())},
				TLS: &v1alpha3.TLSOptions{
					Mode:              "SIMPLE",
					ServerCertificate: filepath.Join(sslDir, fmt.Sprintf("%s-%s", certs.TlsSecretName, "tls.crt")),
					PrivateKey:        filepath.Join(sslDir, fmt.Sprintf("%s-%s", certs.TlsSecretName, "tls.key")),
				},
			})
		}
	}

	for _, publicdomain := range input.PublicDomains {
		if publicdomain.Spec.RequestTLSCert {
			gws.Servers = append(gws.Servers, &v1alpha3.Server{
				Port: &v1alpha3.Port{
					Protocol: "HTTPS",
					Number:   443,
					Name:     publicdomain.Name,
				},
				Hosts: []string{publicdomain.Spec.DomainName},
				TLS: &v1alpha3.TLSOptions{
					Mode:              "SIMPLE",
					ServerCertificate: filepath.Join(sslDir, fmt.Sprintf("%s-%s", fmt.Sprintf("%s-tls-certs", publicdomain.Name), "tls.crt")),
					PrivateKey:        filepath.Join(sslDir, fmt.Sprintf("%s-%s", fmt.Sprintf("%s-tls-certs", publicdomain.Name), "tls.key")),
				},
			})
			if !http80Open {
				// need to open 80 for let's encrypt http challenge
				gws.Servers = append(gws.Servers, &v1alpha3.Server{
					Port: &v1alpha3.Port{
						Protocol: "HTTP",
						Number:   80,
					},
					Hosts: []string{"*"},
				})
			}
		}
	}

	gateway := &istioOutput.Gateway{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Gateway",
			APIVersion: "networking.istio.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      settings.IstioGateway,
			Namespace: settings.RioSystemNamespace,
		},
		Spec: gws,
	}

	output.Gateways[gateway.Name] = gateway
	return nil
}

func GetSSLDir() string {
	return "/etc/istio/ingressgateway-certs"
}
