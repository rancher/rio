package stringers

import (
	"fmt"
	"net"
	"strings"

	"github.com/rancher/wrangler/pkg/kv"
	v1 "k8s.io/api/core/v1"
)

type HostAliasStringer struct {
	v1.HostAlias
}

func (h HostAliasStringer) MaybeString() interface{} {
	return fmt.Sprintf("%s=%s", strings.Join(h.Hostnames, ","), h.IP)
}

func ParseHostAliases(hosts ...string) (result []v1.HostAlias, err error) {
	for _, host := range hosts {
		alias, err := ParseHostAlias(host)
		if err != nil {
			return nil, err
		}
		result = append(result, alias)
	}
	return
}

func ParseHostAlias(host string) (v1.HostAlias, error) {
	hostnames, ip := kv.Split(host, "=")
	if ip == "" {
		return v1.HostAlias{}, fmt.Errorf("%s does not match format host[,host]=ip", host)
	}

	parsed := net.ParseIP(ip)
	if parsed == nil {
		return v1.HostAlias{}, fmt.Errorf("%s is not a valid IP", ip)
	}

	return v1.HostAlias{
		Hostnames: strings.Split(hostnames, ","),
		IP:        ip,
	}, nil
}
