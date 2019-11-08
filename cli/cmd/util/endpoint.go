package util

import (
	"net/url"
	"sort"
)

func NormalizingEndpoints(endpoints []string, suffix string) []string {
	var r []string
	hostNameSeen := map[string]string{}
	for _, e := range endpoints {
		u, _ := url.Parse(e)
		if suffix != "" {
			e = e + suffix
		}
		if u.Scheme == "https" {
			hostNameSeen[u.Hostname()] = e
		} else {
			if _, ok := hostNameSeen[u.Hostname()]; !ok {
				hostNameSeen[u.Hostname()] = e
			}
		}
	}

	for _, v := range hostNameSeen {
		r = append(r, v)
	}
	sort.Strings(r)
	return r
}
