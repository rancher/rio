package populate

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/knative/pkg/apis/istio/v1alpha3"
	"github.com/rancher/rio/features/letsencrypt/controllers/issuer"
	riov1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
)

const (
	sslDir = "/etc/istio/ingressgateway-certs"
)

func populateGateway(systemNamespace string, secret *v1.Secret, publicDomains []*riov1.PublicDomain, output *objectset.ObjectSet) {
	gws := v1alpha3.GatewaySpec{
		Selector: map[string]string{
			"gateway": "external",
		},
	}

	// http port
	port, _ := strconv.ParseInt(settings.DefaultHTTPOpenPort, 10, 0)
	gws.Servers = append(gws.Servers,
		v1alpha3.Server{
			Port: v1alpha3.Port{
				Protocol: "HTTP",
				Number:   int(port),
				Name:     fmt.Sprintf("%v-%v", "http", port),
			},
			Hosts: []string{"*"},
		},
		v1alpha3.Server{
			Port: v1alpha3.Port{
				Protocol: "HTTP2",
				Number:   int(port),
				Name:     fmt.Sprintf("%v-%v", "http2", port),
			},
			Hosts: []string{"*"},
		},
		v1alpha3.Server{
			Port: v1alpha3.Port{
				Protocol: "GRPC",
				Number:   int(port),
				Name:     fmt.Sprintf("%v-%v", "grpc", port),
			},
			Hosts: []string{"*"},
		},
	)

	// https port
	httpsPort, _ := strconv.ParseInt(settings.DefaultHTTPSOpenPort, 10, 0)
	key := fmt.Sprintf("%s-%s", issuer.TLSSecretName, "tls.crt")
	value := fmt.Sprintf("%s-%s", issuer.TLSSecretName, "tls.key")
	if secret != nil && len(secret.Data[key]) > 0 && secret.Annotations["certificate-status"] == "ready" {
		gws.Servers = append(gws.Servers, v1alpha3.Server{
			Port: v1alpha3.Port{
				Protocol: "HTTPS",
				Number:   int(httpsPort),
				Name:     fmt.Sprintf("%v-%v", "https", httpsPort),
			},
			Hosts: []string{fmt.Sprintf("*.%s", settings.ClusterDomain)},
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

	gateway := constructors.NewGateway(systemNamespace, settings.RioGateway, v1alpha3.Gateway{
		Spec: gws,
	})

	output.Add(gateway)
}
