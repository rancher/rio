package populate

import (
	"github.com/rancher/rio/pkg/deploy/istio/input"
	"github.com/rancher/rio/pkg/deploy/istio/output"
	"github.com/rancher/rio/pkg/settings"
)

func populatePorts(input *input.IstioDeployment, output *output.Deployment) error {
	output.Enabled = settings.IstioEnabled.Get() == "true"

	return nil
}
