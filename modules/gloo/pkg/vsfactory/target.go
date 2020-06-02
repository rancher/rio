package vsfactory

import (
	"fmt"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/indexes"
	"github.com/rancher/rio/pkg/serviceports"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/wrangler/pkg/name"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

// the target struct is used to hold information for networking decisions
type target struct {
	Hosts          []string
	Port           int32
	Name           string
	App            string
	Version        string
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

// getTarget returns 1 target for a specified rio service
func (f *VirtualServiceFactory) getTarget(obj *riov1.Service, systemNamespace string) (result target, err error) {
	app, version := services.AppAndVersion(obj)
	result.Name = name.SafeConcatName(app, version)
	result.Namespace = obj.Namespace
	result.Version = version
	result.App = app
	if obj.Status.ComputedWeight != nil {
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

	result.Hosts, err = determineHosts(obj.Status.Endpoints)
	if err != nil {
		return result, err
	}

	if obj.Status.ComputedReplicas != nil { // valid candidate for a service that can be scaled to zero
		result.ScaleIsZero, err = f.isScaleZero(app, version, result.Namespace)
		if err != nil {
			return result, err
		}
		if result.ScaleIsZero {
			if *obj.Status.ComputedReplicas > 0 {
				logrus.Debug("service has ComputedReplicas > 0 but no IP address allocated via k8s endpoints yet")
			}
			result.Port = 80
			result.OriginalTarget = struct {
				Name      string
				Namespace string
			}{Name: obj.Name, Namespace: result.Namespace}
			result.Name = constants.AutoscalerServiceName
			result.Namespace = systemNamespace
		}
	}

	return
}

func (f *VirtualServiceFactory) FindTLS(namespace, app, version string, hostnames []string) (map[string]string, error) {
	result := map[string]string{}

	domains, err := f.clusterDomainCache.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	for _, domain := range domains {
		for _, hostname := range hostnames {
			host, port, err := net.SplitHostPort(hostname)
			if err != nil {
				host = hostname
				port = "443"
			}
			portInt, _ := strconv.Atoi(port)
			if portInt != domain.Spec.HTTPSPort {
				continue
			}
			if strings.HasSuffix(host, domain.Name) {
				if domain.Status.AssignedSecretName != "" {
					result[formatHost(portInt, 443, host)] = domain.Status.AssignedSecretName
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

func (f *VirtualServiceFactory) getTargetsForApp(svcs []*riov1.Service, systemNamespace string) (hostnames []string, targets []target, err error) {
	var (
		seen = map[string]bool{}
	)

	weightSet := false
	for _, svc := range svcs {
		if svc.Spec.Template {
			continue
		}
		target, err := f.getTarget(svc, systemNamespace)
		if err != nil {
			return nil, nil, err
		}
		for _, appEndpoint := range svc.Status.AppEndpoints {
			u, err := url.Parse(appEndpoint)
			if err != nil {
				return nil, nil, err
			}
			hostname := u.Host
			if !seen[hostname] {
				seen[hostname] = true
				hostnames = append(hostnames, hostname)
			}
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
	} else {
		totalWeight := 0
		for i := range targets {
			totalWeight += targets[i].Weight
		}
		addedWeight := 0
		for i := range targets {
			if i == len(targets)-1 {
				targets[i].Weight = 100 - addedWeight
				break
			}
			targets[i].Weight = int(float64(targets[i].Weight) / float64(totalWeight) * 100)
			addedWeight += targets[i].Weight
		}
	}

	return
}

func formatHost(port, defaultPort int, hostname string) string {
	if port == 0 {
		return ""
	}
	if port == defaultPort {
		return hostname
	}
	return fmt.Sprintf("%s:%d", hostname, port)
}

// check if a candidate service is scale zero, does not check if the service CAN be scaled from zero
// returns true if there are NO k8s endpoints for the service allocated yet
func (f *VirtualServiceFactory) isScaleZero(appName, version, namespace string) (bool, error) {
	// for services scaled to zero, ensure endpoint has an IP address
	req, err := labels.NewRequirement("app", selection.Equals, []string{appName})
	if err != nil {
		return false, err
	}
	reqVer, err := labels.NewRequirement("version", selection.Equals, []string{version})
	if err != nil {
		return false, err
	}
	selector := labels.NewSelector().Add(*req)
	selector = selector.Add(*reqVer)
	// use a selector for the app + version endpoint (ie RioApp-v0) versus app endpoint to see if service is up
	k8sEndpoints, err := f.endpoints.List(namespace, selector)
	if err != nil {
		return false, err
	}
	if len(k8sEndpoints) == 0 {
		logrus.Debugf("no corev1.Endpoints found for %s with version %s", appName, version)
		return false, nil
	}
	// check all k8s endpoints to see if any IP addresses exist
	for _, endpoint := range k8sEndpoints {
		if len(endpoint.Subsets) != 0 {
			// ip address allocated for this service, should not be scale zero
			return false, nil
		}
	}
	return true, nil
}

// determineHosts returns a set of hosts based upon the rio services endpoint URLs
func determineHosts(hostNameEndpoints []string) ([]string, error) {
	seen := map[string]bool{}
	hostSet := make([]string, 0, len(hostNameEndpoints))
	for _, endpoint := range hostNameEndpoints {
		u, err := url.Parse(endpoint)
		if err != nil {
			return []string{}, err
		}
		if seen[u.Host] {
			continue
		}
		seen[u.Host] = true

		hostSet = append(hostSet, u.Host)
	}

	sort.Strings(hostSet)
	return hostSet, nil
}
