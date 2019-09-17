package create

import (
	"time"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func populateTimeout(c *Create, service *riov1.Service) error {
	if c.BuildTimeout == "" {
		return nil // rely on tekton to set timeout value
	}
	timeout, err := time.ParseDuration(c.BuildTimeout)
	if err != nil {
		return err
	}
	service.Spec.Build.BuildTimeout = &metav1.Duration{Duration: timeout}
	return nil
}
