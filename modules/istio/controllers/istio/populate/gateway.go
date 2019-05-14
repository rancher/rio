package populate

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/knative/pkg/apis/istio/v1alpha3"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
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

func populateGateway(systemNamespace string, output *objectset.ObjectSet) {
	// Istio Gateway
	gws := v1alpha3.GatewaySpec{
		Selector: map[string]string{
			"gateway": "external",
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
	gateway := constructors.NewGateway(systemNamespace, constants.RioGateway, v1alpha3.Gateway{
		Spec: gws,
	})
	output.Add(gateway)
}
