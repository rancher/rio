package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/ticker"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
)

func (h *handler) getDomain(addrs []adminv1.Address) (string, error) {
	var hosts []string
	var cname bool

	for _, addr := range addrs {
		if addr.Hostname != "" {
			cname = true
			hosts = append(hosts, addr.Hostname)
			break
		}
		if addr.IP != "" {
			hosts = append(hosts, addr.IP)
		}
	}

	if err := h.ensureDomainExists(hosts, cname); err != nil {
		return "", err
	}

	return h.rDNSClient.UpdateDomain(hosts, nil, cname)
}

func (h *handler) ensureDomainExists(hosts []string, cname bool) error {
	domain, err := h.rDNSClient.GetDomain(cname)
	if err != nil && strings.Contains(err.Error(), "forbidden to use") {
		// intentional fall through
	} else if err != nil || domain != nil {
		return err
	}

	if _, err := h.rDNSClient.CreateDomain(hosts, cname); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(h.ctx, time.Second*5)
	defer cancel()
	wait.JitterUntil(func() {
		domain, err = h.rDNSClient.GetDomain(cname)
		if err != nil {
			logrus.Debug("failed to get domain")
		}
	}, time.Second, 1.3, true, ctx.Done())

	if domain == nil {
		return fmt.Errorf("failed to create domain")
	}

	return nil
}

func (h *handler) start() {
	if h.started {
		return
	}

	h.started = true
	go func() {
		for range ticker.Context(h.ctx, 6*time.Hour) {
			logrus.Infof("Renewing rdns domain")
			if err := h.renew(); err != nil {
				logrus.Errorf("failed to renew domain: %v", err)
			}
		}
	}()
}

func (h *handler) renew() error {
	if _, err := h.rDNSClient.RenewDomain(); err != nil {
		return err
	}
	return nil
}
