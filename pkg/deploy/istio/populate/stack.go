package populate

import (
	"fmt"

	"github.com/rancher/rio/pkg/deploy/istio/input"
	"github.com/rancher/rio/pkg/deploy/istio/output"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/stacks"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
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

	ports := []string{fmt.Sprintf("%v:%v", settings.DefaultHTTPOpenPort.Get(), settings.DefaultHTTPOpenPort.Get()), fmt.Sprintf("%v:%v", settings.DefaultHTTPSOpenPort.Get(), settings.DefaultHTTPSOpenPort.Get())}

	portStr, err := json.Marshal(&ports)
	if err != nil {
		return err
	}

	s := &riov1.Stack{
		TypeMeta: v1.TypeMeta{
			Kind:       "Stack",
			APIVersion: "rio.cattle.io/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      settings.IstioStackName,
			Namespace: settings.RioSystemNamespace,
		},
		Spec: riov1.StackSpec{
			Answers: map[string]string{
				"PORTS":               string(portStr),
				"TELEMETRY_NAMESPACE": settings.IstioTelemetryNamespace,
			},
			EnableKubernetesResources: true,
			DisableMesh:               true,
			Template:                  stackContents,
		},
	}

	output.Stacks[s.Name] = s
	return nil
}
