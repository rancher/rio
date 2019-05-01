package stringers

import (
	"fmt"
	"net"
	"sort"

	"github.com/rancher/mapper/convert"
	"github.com/rancher/mapper/mappers"
	"github.com/rancher/wrangler/pkg/kv"
	v1 "k8s.io/api/core/v1"
)

func NewHostAlias(field string) mappers.ObjectsToSlice {
	return mappers.ObjectsToSlice{
		Field: field,
		NewObject: func() mappers.MaybeStringer {
			return &HostAliasStringer{}
		},
		ToObject: func(obj interface{}) (interface{}, error) {
			str, ok := obj.(string)
			if !ok {
				return obj, nil
			}
			objs, err := ParseHostAliases(str)
			if err != nil {
				return nil, err
			}
			if len(objs) == 1 {
				return objs[0], nil
			}

			var result []map[string]interface{}
			for _, obj := range objs {
				newObj, err := convert.EncodeToMap(obj)
				if err != nil {
					return nil, err
				}
				result = append(result, newObj)
			}
			return result, nil
		},
	}
}

type HostAliasStringer struct {
	v1.HostAlias
}

func (h HostAliasStringer) MaybeString() interface{} {
	var ret []string
	for _, host := range h.Hostnames {
		ret = append(ret, fmt.Sprintf("%s:%s", host, h.IP))
	}
	return ret
}

func ParseHostAliases(hosts ...string) (result []v1.HostAlias, err error) {
	hostMap := map[string][]string{}

	for _, host := range hosts {
		hostname, ip := kv.Split(host, ":")
		if ip == "" {
			return nil, fmt.Errorf("%s does not match format host:ip", host)
		}
		parsed := net.ParseIP(ip)
		if parsed == nil {
			return nil, fmt.Errorf("%s is not a valid IP", ip)
		}

		hostMap[ip] = append(hostMap[ip], hostname)
	}

	for ip, hosts := range hostMap {
		result = append(result, v1.HostAlias{
			IP:        ip,
			Hostnames: hosts,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].IP < result[j].IP
	})

	return result, nil
}
