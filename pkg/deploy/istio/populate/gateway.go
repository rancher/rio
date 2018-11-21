package populate

import (
	"fmt"
	"path/filepath"
	"strconv"

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

	gws := v1alpha3.GatewaySpec{
		Selector: settings.IstioGatewaySelector,
	}

	// http port
	port, _ := strconv.ParseInt(settings.DefaultHTTPOpenPort.Get(), 10, 0)
	gws.Servers = append(gws.Servers, &v1alpha3.Server{
		Port: &v1alpha3.Port{
			Protocol: "HTTP",
			Number:   uint32(port),
			Name:     fmt.Sprintf("%v-%v", "http", port),
		},
		Hosts: []string{"*"},
	})

	// https port
	sslDir := GetSSLDir()
	httpsPort, _ := strconv.ParseInt(settings.DefaultHTTPSOpenPort.Get(), 10, 0)
	key := fmt.Sprintf("%s-%s", certs.TlsSecretName, "tls.crt")
	value := fmt.Sprintf("%s-%s", certs.TlsSecretName, "tls.key")
	if input.Secret != nil && len(input.Secret.Data[key]) > 0 && input.Secret.Annotations["certificate-status"] == "ready" {
		gws.Servers = append(gws.Servers, &v1alpha3.Server{
			Port: &v1alpha3.Port{
				Protocol: "HTTPS",
				Number:   uint32(httpsPort),
				Name:     fmt.Sprintf("%v-%v", "https", httpsPort),
			},
			Hosts: []string{fmt.Sprintf("*.%s", settings.ClusterDomain.Get())},
			TLS: &v1alpha3.TLSOptions{
				Mode:              "SIMPLE",
				ServerCertificate: filepath.Join(sslDir, key),
				PrivateKey:        filepath.Join(sslDir, value),
			},
		})
	}

	for _, publicdomain := range input.PublicDomains {
		key := fmt.Sprintf("%s-%s", fmt.Sprintf("%s-tls-certs", publicdomain.Name), "tls.crt")
		value := fmt.Sprintf("%s-%s", fmt.Sprintf("%s-tls-certs", publicdomain.Name), "tls.key")
		if input.Secret != nil && len(input.Secret.Data[key]) > 0 && publicdomain.Annotations["certificate-status"] == "ready" {
			gws.Servers = append(gws.Servers, &v1alpha3.Server{
				Port: &v1alpha3.Port{
					Protocol: "HTTPS",
					Number:   443,
					Name:     publicdomain.Name,
				},
				Hosts: []string{publicdomain.Spec.DomainName},
				TLS: &v1alpha3.TLSOptions{
					Mode:              "SIMPLE",
					ServerCertificate: filepath.Join(sslDir, key),
					PrivateKey:        filepath.Join(sslDir, value),
				},
			})
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
