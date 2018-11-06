package populate

import (
	"fmt"
	"strings"

	"github.com/rancher/rio/pkg/deploy/istio/input"
	"github.com/rancher/rio/pkg/deploy/istio/output"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/stacks"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
)

var (
	stackContents = string(stacks.MustAsset("stacks/istio-stack.yml"))
)

func populateStack(input *input.IstioDeployment, output *output.Deployment) error {
	if !output.Enabled {
		return nil
	}

	var ports []string
	for _, port := range output.Ports {
		p := strings.SplitN(port, "/", 2)[0]
		ports = append(ports, fmt.Sprintf("%v:%v", p, p))
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
			Name:      settings.IstioStackName,
			Namespace: settings.RioSystemNamespace,
		},
		Spec: v1beta1.StackSpec{
			Answers: map[string]string{
				"PORTS": string(portStr),
			},
			EnableKubernetesResources: true,
			DisableMesh:               true,
			Template:                  stackContents,
		},
	}

	output.Stacks[s.Name] = s
	return nil
}
