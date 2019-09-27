package vsfactory

import (
	"fmt"
	"net"
	"net/url"
	"sort"
	"strings"

	"github.com/rancher/rio/modules/service/controllers/service/populate/serviceports"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/indexes"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/wrangler/pkg/name"
	"k8s.io/apimachinery/pkg/labels"
)

type target struct {
	Hosts          []string
	Port           int32
	Name           string
	Namespace      string
	Weight         int
	ScaleIsZero    bool
	OriginalTarget struct {
		Name      string
		Namespace string
	}
}

func (t target) valid() bool {
	return t.Port != 0 && len(t.Hosts) > 0
}

func getTarget(obj *riov1.Service, systemNamespace string) (result target, err error) {
	app, version := services.AppAndVersion(obj)
	result.Name = name.SafeConcatName(app, version)
	result.Namespace = obj.Namespace
	if obj.Spec.RolloutConfig != nil && obj.Status.ComputedWeight != nil {
		result.Weight = *obj.Status.ComputedWeight
	} else if obj.Spec.Weight != nil {
		result.Weight = *obj.Spec.Weight
	}

	for _, port := range serviceports.ContainerPorts(obj) {
		if port.IsExposed() && port.IsHTTP() {
			result.Port = port.Port
			continue
		}
	}

	if obj.Status.ComputedReplicas != nil && *obj.Status.ComputedReplicas == 0 {
		result.ScaleIsZero = true
		result.Port = 80
		result.OriginalTarget = struct {
			Name      string
			Namespace string
		}{Name: obj.Name, Namespace: result.Namespace}
		result.Name = constants.AutoscalerServiceName
		result.Namespace = systemNamespace
	}

	seen := map[string]bool{}
	for _, endpoint := range obj.Status.Endpoints {
		u, err := url.Parse(endpoint)
		if err != nil {
			return result, err
		}
		if seen[u.Host] {
			continue
		}
		seen[u.Host] = true

		result.Hosts = append(result.Hosts, u.Host)
	}

	sort.Strings(result.Hosts)
	return
}

func (f *VirtualServiceFactory) findTLS(namespace, app, version string, hostnames []string) (map[string]string, error) {
	result := map[string]string{}

	domains, err := f.clusterDomainCache.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	for _, domain := range domains {
		for _, hostname := range hostnames {
			host, _, err := net.SplitHostPort(hostname)
			if err != nil {
				host = hostname
			}
			if strings.HasSuffix(host, domain.Name) {
				if domain.Status.AssignedSecretName != "" {
					result[host] = domain.Status.AssignedSecretName
				}
			}
		}
	}

	key := fmt.Sprintf("%s/%s", namespace, app)
	if version != "" {
		key = fmt.Sprintf("%s/%s/%s", namespace, app, version)
	}

	pds, err := f.publicDomainCache.GetByIndex(indexes.PublicDomainByTarget, key)
	if err != nil {
		return nil, err
	}

	for _, pd := range pds {
		if pd.Status.AssignedSecretName != "" {
			result[pd.Name] = pd.Status.AssignedSecretName
		}
	}

	return result, nil
}

func getTargetsForApp(svcs []*riov1.Service, systemNamespace string) (hostnames []string, targets []target, err error) {
	var (
		seen = map[string]bool{}
	)

	weightSet := false
	for _, svc := range svcs {
		target, err := getTarget(svc, systemNamespace)
		if err != nil {
			return nil, nil, err
		}
		for _, appEndpoint := range svc.Status.AppEndpoints {
			u, err := url.Parse(appEndpoint)
			if err != nil {
				return nil, nil, err
			}
			hostname := u.Host
			if seen[hostname] {
				continue
			}
			seen[hostname] = true
			hostnames = append(hostnames, hostname)
		}
		if target.Weight != 0 {
			weightSet = true
		}
		targets = append(targets, target)
	}

	if !weightSet {
		for i := range targets {
			targets[i].Weight = 1
		}
	}

	return
}
