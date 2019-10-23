package stringers

import (
	"fmt"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
)

type PodDNSConfigOptionStringer struct {
	riov1.PodDNSConfigOption
}

func (p PodDNSConfigOptionStringer) MaybeString() interface{} {
	if p.Value == nil {
		return p.Name
	}
	return fmt.Sprintf("%s:%s", p.Name, *p.Value)
}

func ParseDNSOptions(options ...string) (result []riov1.PodDNSConfigOption, err error) {
	for _, opt := range options {
		dns, err := ParseDNSOption(opt)
		if err != nil {
			return nil, err
		}
		result = append(result, dns)
	}
	return
}

func ParseDNSOption(option string) (riov1.PodDNSConfigOption, error) {
	k, v := kv.Split(option, ":")
	podDNSOpt := riov1.PodDNSConfigOption{
		Name: k,
	}
	if v != "" {
		podDNSOpt.Value = &v
	}

	return podDNSOpt, nil
}
