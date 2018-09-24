package populate

import (
	"fmt"

	"github.com/rancher/rio/pkg/deploy/istio/input"
	"github.com/rancher/rio/pkg/deploy/istio/output"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/stacks"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
)

var (
	proxyStackContents = string(stacks.MustAsset("stacks/istio-gw-stack.yml"))
)

func populateProxy(input *input.IstioDeployment, output *output.Deployment) error {
	if !output.Enabled {
		return nil
	}

	var ports []string
	for _, port := range output.Ports {
		ports = append(ports, fmt.Sprintf("%d:%d", port, port))
	}

	portStr, err := json.Marshal(&ports)
	if err != nil {
		return err
	}

	s := &v1beta1.Stack{
		TypeMeta: v1.TypeMeta{
			Kind:       "Stack",
			APIVersion: "rio.cattle.io/v1beta1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "istio-proxy",
			Namespace: settings.RioSystemNamespace,
		},
		Spec: v1beta1.StackSpec{
			Answers: map[string]string{
				"PORTS": string(portStr),
			},
			DisableMesh: true,
			Template:    proxyStackContents,
		},
	}

	output.Stacks[s.Name] = s
	return nil
}
