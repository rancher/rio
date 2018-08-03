package create

import (
	"time"

	"github.com/rancher/rio/types/client/rio/v1beta1"
)

func populateHealthCheck(c *Create, service *client.Service) error {
	var err error

	hc := &client.HealthConfig{
		HealthyThreshold: int64(c.HealthRetries),
	}

	hc.InitialDelaySeconds, err = ParseDurationUnit(c.HealthStartPeriod, "--health-start-period", time.Second)
	if err != nil {
		return err
	}

	hc.IntervalSeconds, err = ParseDurationUnit(c.HealthInterval, "--health-interval", time.Second)
	if err != nil {
		return err
	}

	if len(c.HealthCmd) > 0 {
		hc.Test = []string{"CMD-SHELL", c.HealthCmd}
	}

	if len(c.HealthURL) > 0 {
		hc.Test = []string{c.HealthCmd}
	}

	hc.TimeoutSeconds, err = ParseDurationUnit(c.HealthTimeout, "--health-timeout", time.Second)

	if len(c.HealthCmd) > 0 {
		service.Healthcheck = hc
	}

	return err
}
