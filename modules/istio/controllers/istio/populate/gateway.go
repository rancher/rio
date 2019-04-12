package populate

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/knative/pkg/apis/istio/v1alpha3"
	riov1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
)

const (
	sslDir  = "/etc/istio/ingressgateway-certs"
	tlsKey  = "tls.key"
	tlsCert = "tls.crt"
)

var (
	supportedProtocol = []v1alpha3.PortProtocol{
		v1alpha3.ProtocolHTTP,
		//v1alpha3.ProtocolHTTPS,
		//v1alpha3.ProtocolTCP,
		//v1alpha3.ProtocolGRPC,
		//v1alpha3.ProtocolHTTP2,
	}
)

func populateGateway(systemNamespace string, clusterDomain *riov1.ClusterDomain, wildcardSecret *v1.Secret, publicDomains []*riov1.PublicDomain, publicDomainSecrets map[string]*v1.Secret, output *objectset.ObjectSet) {
	gws := v1alpha3.GatewaySpec{
		Selector: map[string]string{
			"gateway": "external",
		},
	}

	// http port
	port, _ := strconv.ParseInt(settings.DefaultHTTPOpenPort, 10, 0)
	for _, protocol := range supportedProtocol {
		gws.Servers = append(gws.Servers, v1alpha3.Server{
			Port: v1alpha3.Port{
				Protocol: protocol,
				Number:   int(port),
				Name:     fmt.Sprintf("%v-%v", strings.ToLower(string(protocol)), port),
			},
			Hosts: []string{"*"},
		})
	}

	// https port
	httpsPort, _ := strconv.ParseInt(settings.DefaultHTTPSOpenPort, 10, 0)
	if wildcardSecret != nil {
		if len(wildcardSecret.Data[tlsCert]) > 0 {
			gws.Servers = append(gws.Servers, v1alpha3.Server{
				Port: v1alpha3.Port{
					Protocol: v1alpha3.ProtocolHTTPS,
					Number:   int(httpsPort),
					Name:     fmt.Sprintf("%v-%v", "https", httpsPort),
				},
				Hosts: []string{fmt.Sprintf("*.%s", clusterDomain.Status.ClusterDomain)},
				TLS: &v1alpha3.TLSOptions{
					Mode:              v1alpha3.TLSModeSimple,
					ServerCertificate: filepath.Join(sslDir, tlsCert),
					PrivateKey:        filepath.Join(sslDir, tlsKey),
				},
			})
		}
	}

	for _, publicdomain := range publicDomains {
		key := fmt.Sprintf("%s/%s", publicdomain.Namespace, publicdomain.Name)
		secret := publicDomainSecrets[key]
		if secret != nil && len(secret.Data[key]) > 0 {
			gws.Servers = append(gws.Servers, v1alpha3.Server{
				Port: v1alpha3.Port{
					Protocol: v1alpha3.ProtocolHTTPS,
					Number:   443,
					Name:     publicdomain.Name,
				},
				Hosts: []string{publicdomain.Spec.DomainName},
				TLS: &v1alpha3.TLSOptions{
					Mode:              v1alpha3.TLSModeSimple,
					ServerCertificate: filepath.Join(sslDir, tlsCert),
					PrivateKey:        filepath.Join(sslDir, tlsKey),
				},
			})
		}
	}

	gateway := constructors.NewGateway(systemNamespace, settings.RioGateway, v1alpha3.Gateway{
		Spec: gws,
	})

	output.Add(gateway)
}
