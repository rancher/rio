package populate

import (
	"fmt"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/stacks"
	"github.com/rancher/wrangler/pkg/objectset"
	"k8s.io/apimachinery/pkg/util/json"
)

var (
	stackContents = string(stacks.MustAsset("stacks/istio-stack.yaml"))
)

func PopulateStack(systemNamespace string, output *objectset.ObjectSet) error {
	ports := []string{
		fmt.Sprintf("%v:%v", settings.DefaultHTTPOpenPort, settings.DefaultHTTPOpenPort),
		fmt.Sprintf("%v:%v", settings.DefaultHTTPSOpenPort, settings.DefaultHTTPSOpenPort),
	}

	portStr, err := json.Marshal(&ports)
	if err != nil {
		return err
	}

	s := riov1.NewStack(systemNamespace, settings.IstioStackName, riov1.Stack{
		Spec: riov1.StackSpec{
			Answers: map[string]string{
				"PORTS":               string(portStr),
				"TELEMETRY_NAMESPACE": settings.IstioTelemetry,
				"PILOT_ADDRESS":       fmt.Sprintf("%s.%s", "istio-pilot", settings.IstioStackName),
			},
			EnableKubernetesResources: true,
			DisableMesh:               true,
			Template:                  stackContents,
		},
	})

	output.Add(s)
	return nil
}
