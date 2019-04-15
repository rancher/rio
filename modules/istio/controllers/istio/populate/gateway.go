package populate

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rancher/rio/modules/system/features/letsencrypt/pkg/issuers"

	"github.com/knative/pkg/apis/istio/v1alpha3"
	projectv1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
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

func populateGateway(systemNamespace string, clusterDomain *projectv1.ClusterDomain, secret *v1.Secret, publicDomains []*riov1.PublicDomain, output *objectset.ObjectSet) {
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

	// wildcards
	httpsPort, _ := strconv.ParseInt(settings.DefaultHTTPSOpenPort, 10, 0)

	if secret != nil {
		crtFile := fmt.Sprintf("%s-%s-tls.crt", systemNamespace, issuers.RioWildcardCerts)
		keyPath, crtPath := keyCertPath(systemNamespace, issuers.RioWildcardCerts)
		if len(secret.Data[crtFile]) > 0 {
			gws.Servers = append(gws.Servers, v1alpha3.Server{
				Port: v1alpha3.Port{
					Protocol: v1alpha3.ProtocolHTTPS,
					Number:   int(httpsPort),
					Name:     fmt.Sprintf("%v-%v", "https", httpsPort),
				},
				Hosts: []string{fmt.Sprintf("*.%s", clusterDomain.Status.ClusterDomain)},
				TLS: &v1alpha3.TLSOptions{
					Mode:              v1alpha3.TLSModeSimple,
					ServerCertificate: crtPath,
					PrivateKey:        keyPath,
				},
			})
		}
	}

	for _, publicdomain := range publicDomains {
		crtFile := fmt.Sprintf("%s-%s-tls.crt", publicdomain.Spec.SecretRef.Namespace, publicdomain.Spec.SecretRef.Name)
		keyPath, crtPath := keyCertPath(publicdomain.Spec.SecretRef.Namespace, publicdomain.Spec.SecretRef.Name)
		if len(secret.Data[crtFile]) > 0 {
			gws.Servers = append(gws.Servers, v1alpha3.Server{
				Port: v1alpha3.Port{
					Protocol: v1alpha3.ProtocolHTTPS,
					Number:   443,
					Name:     publicdomain.Name,
				},
				Hosts: []string{publicdomain.Spec.DomainName},
				TLS: &v1alpha3.TLSOptions{
					Mode:              v1alpha3.TLSModeSimple,
					ServerCertificate: crtPath,
					PrivateKey:        keyPath,
				},
			})
		}
	}

	gateway := constructors.NewGateway(systemNamespace, settings.RioGateway, v1alpha3.Gateway{
		Spec: gws,
	})

	output.Add(gateway)
}

func keyCertPath(namespace, name string) (string, string) {
	key := fmt.Sprintf("%s-%s-tls.key", namespace, name)
	cert := fmt.Sprintf("%s-%s-tls.crt", namespace, name)
	return filepath.Join(sslDir, key), filepath.Join(sslDir, cert)
}
