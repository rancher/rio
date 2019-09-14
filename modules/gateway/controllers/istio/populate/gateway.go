package populate

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/knative/pkg/apis/istio/v1alpha3"
	v1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
)

var (
	supportedProtocol = []v1alpha3.PortProtocol{
		v1alpha3.ProtocolHTTP,
		//v1alpha3.ProtocolTCP,
		//v1alpha3.ProtocolGRPC,
		//v1alpha3.ProtocolHTTP2,
	}
)

func Gateway(systemNamespace, clusterDomain, certName string, publicdomains []*v1.PublicDomain, output *objectset.ObjectSet) {
	// Istio Gateway
	gws := v1alpha3.GatewaySpec{
		Selector: map[string]string{
			"app": constants.GatewayName,
		},
	}

	// http port
	port, _ := strconv.ParseInt(constants.DefaultHTTPOpenPort, 10, 0)
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

	if constants.InstallMode != constants.InstallModeIngress {
		// https port
		if clusterDomain != "" {
			httpsPort, _ := strconv.ParseInt(constants.DefaultHTTPSOpenPort, 10, 0)
			gws.Servers = append(gws.Servers, v1alpha3.Server{
				Port: v1alpha3.Port{
					Protocol: v1alpha3.ProtocolHTTPS,
					Number:   int(httpsPort),
					Name:     fmt.Sprintf("%v-%v", strings.ToLower(string(v1alpha3.ProtocolHTTPS)), httpsPort),
				},
				Hosts: []string{"*"},
				TLS: &v1alpha3.TLSOptions{
					Mode:           v1alpha3.TLSModeSimple,
					CredentialName: certName,
				},
			})
		}

		for _, pd := range publicdomains {
			httpsPort, _ := strconv.ParseInt(constants.DefaultHTTPSOpenPort, 10, 0)
			gws.Servers = append(gws.Servers, v1alpha3.Server{
				Port: v1alpha3.Port{
					Protocol: v1alpha3.ProtocolHTTPS,
					Number:   int(httpsPort),
					Name:     fmt.Sprintf("%v-%v-%v", pd.Name, strings.ToLower(string(v1alpha3.ProtocolHTTPS)), httpsPort),
				},
				Hosts: []string{pd.Spec.DomainName},
				TLS: &v1alpha3.TLSOptions{
					Mode:           v1alpha3.TLSModeSimple,
					CredentialName: pd.Spec.SecretRef.Name,
				},
			})
		}
	}

	gateway := constructors.NewGateway(systemNamespace, constants.RioGateway, v1alpha3.Gateway{
		Spec: gws,
	})
	output.Add(gateway)
}
