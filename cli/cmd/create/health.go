package create

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/mattn/go-shellwords"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func populateHealthCheck(c *Create, service *riov1.Service) error {
	if c.HealthURL == "" && c.HealthCmd == "" {
		return nil
	}

	hc := v1.Probe{
		FailureThreshold: int32(c.HealthFailureThreshold),
		SuccessThreshold: int32(c.HealthSuccessThreshold),
	}

	if c.HealthInitialDelay != "" {
		delay, err := time.ParseDuration(c.HealthInitialDelay)
		if err != nil {
			return err
		}

		hc.InitialDelaySeconds = int32(delay.Seconds())
	}

	if c.HealthInterval != "" {
		interval, err := time.ParseDuration(c.HealthInterval)
		if err != nil {
			return err
		}

		hc.PeriodSeconds = int32(interval.Seconds())
	}

	if c.HealthTimeout != "" {
		timeout, err := time.ParseDuration(c.HealthTimeout)
		if err != nil {
			return err
		}

		hc.TimeoutSeconds = int32(timeout.Seconds())
	}

	if len(c.HealthCmd) > 0 {
		words, err := shellwords.Parse(c.HealthCmd)
		if err != nil {
			return err
		}
		hc.Handler.Exec = &v1.ExecAction{
			Command: words,
		}
	}

	if len(c.HealthURL) > 0 {
		u, err := url.Parse(c.HealthURL)
		if err != nil {
			return err
		}

		portStr := u.Port()
		if portStr == "" {
			return fmt.Errorf("missing port in health-url %s", c.HealthURL)
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return err
		}

		if u.Scheme == "tcp" {
			hc.TCPSocket = &v1.TCPSocketAction{
				Port: intstr.FromInt(port),
			}
		} else {
			hc.HTTPGet = &v1.HTTPGetAction{
				Port: intstr.FromInt(port),
				Host: u.Host,
				Path: u.Path,
			}

			for key, value := range c.HealthHeader {
				hc.HTTPGet.HTTPHeaders = append(hc.HTTPGet.HTTPHeaders, v1.HTTPHeader{
					Name:  key,
					Value: value,
				})
			}

			if u.Scheme == "http" {
				hc.HTTPGet.Scheme = v1.URISchemeHTTP
			} else if u.Scheme == "https" {
				hc.HTTPGet.Scheme = v1.URISchemeHTTPS
			} else {
				return fmt.Errorf("invalid scheme in health-url %s: %s", c.HealthURL, u.Scheme)
			}

		}
	}

	service.Spec.LivenessProbe = &hc
	service.Spec.ReadinessProbe = &hc

	return nil
}
