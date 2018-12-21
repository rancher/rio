package populate

import (
	"fmt"

	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/stacks"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"k8s.io/apimachinery/pkg/util/json"
)

var (
	stackContents = string(stacks.MustAsset("stacks/istio-stack.yaml"))
)

func PopulateStack(output *objectset.ObjectSet) error {
	ports := []string{
		fmt.Sprintf("%v:%v", settings.DefaultHTTPOpenPort.Get(), settings.DefaultHTTPOpenPort.Get()),
		fmt.Sprintf("%v:%v", settings.DefaultHTTPSOpenPort.Get(), settings.DefaultHTTPSOpenPort.Get()),
	}

	portStr, err := json.Marshal(&ports)
	if err != nil {
		return err
	}

	s := riov1.NewStack(settings.RioSystemNamespace, settings.IstioStackName, riov1.Stack{
		Spec: riov1.StackSpec{
			Answers: map[string]string{
				"PORTS":               string(portStr),
				"TELEMETRY_NAMESPACE": settings.IstioTelemetryNamespace,
			},
			EnableKubernetesResources: true,
			DisableMesh:               true,
			Template:                  stackContents,
		},
	})

	output.Add(s)
	return nil
}
