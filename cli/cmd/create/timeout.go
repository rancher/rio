package create

import (
	"time"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/services"
)

func (c *Create) setBuildOrImage(imageName string, spec *riov1.ServiceSpec) error {
	if services.IsRepo(imageName) {
		spec.ImageBuild = &riov1.ImageBuildSpec{
			Branch:                 c.BuildBranch,
			Dockerfile:             c.BuildDockerfile,
			Context:                c.BuildContext,
			Revision:               c.BuildRevision,
			WebhookSecretName:      c.BuildWebhookSecret,
			CloneSecretName:        c.BuildCloneSecret,
			ImageName:              c.BuildImageName,
			PushRegistry:           c.BuildRegistry,
			PushRegistrySecretName: c.BuildDockerPushSecret,
			Repo:                   imageName,
			PR:                     c.BuildPr,
		}

		if c.BuildTimeout != "" {
			timeout, err := time.ParseDuration(c.BuildTimeout)
			if err != nil {
				return err
			}
			sec := int(timeout / time.Second)
			spec.ImageBuild.TimeoutSeconds = &sec
		}
	} else {
		spec.Image = imageName
	}

	return nil
}
