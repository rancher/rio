package create

import (
	"errors"
	"fmt"
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
			if c.BuildWebhookSecret == "" {
				fmt.Println("Warning: tag mode only works with a webhook")
			}
		}
		if c.BuildPr == true {
			if c.Template == false {
				return errors.New("build-pr is only compatible with template mode")
			}
			if c.BuildWebhookSecret == "" {
				fmt.Println("Warning: build-pr only works with a webhook")
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
