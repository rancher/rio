package populate

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/knative/pkg/apis/istio/v1alpha3"
	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/features/letsencrypt/controllers/issuer"
	"github.com/rancher/rio/pkg/settings"
	v1alpha3client "github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3"
	riov1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	"k8s.io/api/core/v1"
)

const (
	sslDir = "/etc/istio/ingressgateway-certs"
)

func populateGateway(secret *v1.Secret, publicDomains []*riov1.PublicDomain, output *objectset.ObjectSet) {
	gws := v1alpha3.GatewaySpec{
		Selector: settings.IstioGatewaySelector,
	}

	// http port
	port, _ := strconv.ParseInt(settings.DefaultHTTPOpenPort.Get(), 10, 0)
	gws.Servers = append(gws.Servers, v1alpha3.Server{
		Port: v1alpha3.Port{
			Protocol: "HTTP",
			Number:   int(port),
			Name:     fmt.Sprintf("%v-%v", "http", port),
		},
		Hosts: []string{"*"},
	})

	// https port
	httpsPort, _ := strconv.ParseInt(settings.DefaultHTTPSOpenPort.Get(), 10, 0)
	key := fmt.Sprintf("%s-%s", issuer.TLSSecretName, "tls.crt")
	value := fmt.Sprintf("%s-%s", issuer.TLSSecretName, "tls.key")
	if secret != nil && len(secret.Data[key]) > 0 && secret.Annotations["certificate-status"] == "ready" {
		gws.Servers = append(gws.Servers, v1alpha3.Server{
			Port: v1alpha3.Port{
				Protocol: "HTTPS",
				Number:   int(httpsPort),
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

	for _, publicdomain := range publicDomains {
		key := fmt.Sprintf("%s-%s", fmt.Sprintf("%s-tls-certs", publicdomain.Name), "tls.crt")
		value := fmt.Sprintf("%s-%s", fmt.Sprintf("%s-tls-certs", publicdomain.Name), "tls.key")
		if secret != nil && len(secret.Data[key]) > 0 && publicdomain.Annotations["certificate-status"] == "ready" {
			gws.Servers = append(gws.Servers, v1alpha3.Server{
				Port: v1alpha3.Port{
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

	gateway := v1alpha3client.NewGateway(settings.RioSystemNamespace, settings.IstioGateway, v1alpha3client.Gateway{
		Spec: gws,
	})

	output.Add(gateway)
}
