package create

import (
	"time"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/pretty/objectmappers"
)

func populateHealthCheck(c *Create, service *riov1.Service) error {
	var err error

	hc := &riov1.HealthConfig{
		HealthyThreshold:   c.HealthRetries,
		UnhealthyThreshold: c.UnhealthyRetries,
	}

	hc.InitialDelaySeconds, err = objectmappers.ParseDurationUnit(c.HealthStartPeriod, "--health-start-period", time.Second)
	if err != nil {
		return err
	}

	hc.IntervalSeconds, err = objectmappers.ParseDurationUnit(c.HealthInterval, "--health-interval", time.Second)
	if err != nil {
		return err
	}

	if len(c.HealthCmd) > 0 {
		hc.Test = []string{"CMD-SHELL", c.HealthCmd}
	}

	if len(c.HealthURL) > 0 {
		hc.Test = []string{c.HealthURL}
	}

	hc.TimeoutSeconds, err = objectmappers.ParseDurationUnit(c.HealthTimeout, "--health-timeout", time.Second)
	if err != nil {
		return err
	}

	if len(c.HealthCmd) > 0 || len(c.HealthURL) > 0 {
		service.Spec.Healthcheck = hc
	}

	return populateReadyCheck(c, service)
}

func populateReadyCheck(c *Create, service *riov1.Service) error {
	var err error

	hc := &riov1.HealthConfig{
		HealthyThreshold:   c.ReadyRetries,
		UnhealthyThreshold: c.UnreadyRetries,
	}

	hc.InitialDelaySeconds, err = objectmappers.ParseDurationUnit(c.ReadyStartPeriod, "--ready-start-period", time.Second)
	if err != nil {
		return err
	}

	hc.IntervalSeconds, err = objectmappers.ParseDurationUnit(c.ReadyInterval, "--ready-interval", time.Second)
	if err != nil {
		return err
	}

	if len(c.ReadyCmd) > 0 {
		hc.Test = []string{"CMD-SHELL", c.ReadyCmd}
	}

	if len(c.ReadyURL) > 0 {
		hc.Test = []string{c.ReadyURL}
	}

	hc.TimeoutSeconds, err = objectmappers.ParseDurationUnit(c.ReadyTimeout, "--ready-timeout", time.Second)

	if len(c.ReadyCmd) > 0 || len(c.ReadyURL) > 0 {
		service.Spec.Readycheck = hc
	}

	return err
}
