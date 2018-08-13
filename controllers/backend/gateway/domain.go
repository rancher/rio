package gateway

import (
	"fmt"
	"time"

	"sort"

	"github.com/rancher/norman/types/slice"
	"github.com/rancher/rio/pkg/settings"
)

var (
	refreshInterval = 5 * time.Minute
)

func (g *Controller) renew() error {
	_, err := g.rdnsClient.RenewDomain()
	return err
}

func (g *Controller) updateDomain(ips []string) error {
	var (
		fqdn string
		err  error
	)

	if len(ips) == 0 {
		return nil
	}

	sort.Strings(ips)
	if slice.StringsEqual(g.previousIPs, ips) {
		return nil
	}

	if len(ips) == 1 && ips[0] == "127.0.0.1" {
		fqdn = nip(ips[0])
	} else {
		_, fqdn, err = g.rdnsClient.ApplyDomain(ips)
		if err != nil {
			if settings.ClusterDomain.Get() == "" {
				settings.ClusterDomain.Set(nip(ips[0]))
			}
			return err
		}
	}

	settings.ClusterDomain.Set(fqdn)
	g.previousIPs = ips
	return nil
}

func nip(ip string) string {
	return fmt.Sprintf("%s.nip.io", ip)
}
