package populate

import (
	"encoding/json"
	"fmt"

	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/pkg/systemstack"
	"github.com/rancher/rio/stacks"
	"github.com/rancher/wrangler/pkg/apply"
)

var (
	stackContents = string(stacks.MustAsset("stacks/istio-stack.yaml"))
)

func IstioDeploy(systemNamespace string, apply apply.Apply) error {
	ports := []string{
		fmt.Sprintf("%v:%v,http2", settings.DefaultHTTPOpenPort, settings.DefaultHTTPOpenPort),
		fmt.Sprintf("%v:%v,https", settings.DefaultHTTPSOpenPort, settings.DefaultHTTPSOpenPort),
	}

	portStr, err := json.Marshal(&ports)
	if err != nil {
		return err
	}
	answers := map[string]string{
		"PORTS":     string(portStr),
		"NAMESPACE": systemNamespace,
	}
	istioStack := systemstack.NewSystemStack(apply, systemNamespace, "istio")
	return istioStack.Deploy(answers)
}
