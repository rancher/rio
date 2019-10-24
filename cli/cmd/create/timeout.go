package create

import (
	"time"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/services"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Create) setBuildOrImage(imageName string, spec *riov1.ServiceSpec) error {
	if services.IsRepo(imageName) {
		spec.ImageBuild = &riov1.ImageBuildSpec{
			Branch:                 c.BuildBranch,
			Dockerfile:             c.BuildDockerfile,
			Context:                c.BuildContext,
			Revision:               c.BuildRevision,
			WebhookSecretName:      c.BuildWebhookSecret,
			CloneSecretName:        c.CloneGitSecret,
			ImageName:              c.BuildImageName,
			PushRegistry:           c.BuildRegistry,
			PushRegistrySecretName: c.BuildDockerPushSecret,
			Repo:                   imageName,
			PR:                     c.BuildPr,
		}
		spec.Template = c.BuildTemplate

		if c.BuildTimeout != "" {
			timeout, err := time.ParseDuration(c.BuildTimeout)
			if err != nil {
				return err
			}
			spec.ImageBuild.Timeout = &v1.Duration{Duration: timeout}
		}
	} else {
		spec.Image = imageName
	}

	return nil
}
