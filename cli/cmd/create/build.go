package create

import (
	"errors"
	"time"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/services"
)

func (c *Create) setBuildOrImage(imageName string, spec *riov1.ServiceSpec) error {
	if services.IsRepo(imageName) {
		if c.BuildTag == true {
			if c.BuildBranch != "master" {
				return errors.New("build-branch and build-tag cannot both be set, as build-tag will deploy tags from every branch")
			}
			if c.BuildPr == true {
				return errors.New("build-tag and build-pr cannot both be turned on")
			}
		}

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
			Tag:                    c.BuildTag,
			TagIncludeRegexp:       c.BuildTagInclude,
			TagExcludeRegexp:       c.BuildTagExclude,
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
