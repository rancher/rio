package build

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/deploymentstatus"
	"github.com/rancher/rio/cli/pkg/progress"
	"github.com/rancher/rio/pkg/config"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func EnableBuildAndWait(ctx *clicontext.CLIContext) error {
	cm, err := ctx.Core.ConfigMaps(ctx.SystemNamespace).Get(config.ConfigName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	conf, err := config.FromConfigMap(cm)
	if err != nil {
		return err
	}

	if conf.Features == nil {
		conf.Features = map[string]config.FeatureConfig{}
	}

	t := true
	f := conf.Features["build"]
	f.Enabled = &t
	conf.Features["build"] = f

	cm, err = config.SetConfig(cm, conf)
	if err != nil {
		return err
	}
	if _, err := ctx.Core.ConfigMaps(ctx.SystemNamespace).Update(cm); err != nil {
		return err
	}

	writer := progress.NewWriter()
	for {
		deployment, err := ctx.Apps.Deployments(ctx.SystemNamespace).Get("buildkitd", metav1.GetOptions{})
		if err != nil && !errors.IsNotFound(err) {
			return err
		}

		if deployment == nil || !deploymentstatus.IsReady(deployment.Status) {
			writer.Display("Waiting for buildkitd to start", 2)
			continue
		}
		break
	}

	return nil
}
