package pod

import (
	"fmt"
	"strings"

	"github.com/rancher/norman/pkg/kv"
	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/features/stack/controllers/service/populate/container"
	"github.com/rancher/rio/features/stack/controllers/service/populate/podvolume"
	"github.com/rancher/rio/features/stack/controllers/service/populate/rbac"
	"github.com/rancher/rio/features/stack/controllers/service/populate/servicelabels"
	sidekick2 "github.com/rancher/rio/features/stack/controllers/service/populate/sidekick"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Populate(stack *riov1.Stack, configsByName map[string]*riov1.Config, volumeDefs map[string]*riov1.Volume, service *riov1.Service, os *objectset.ObjectSet) v1.PodTemplateSpec {
	podSpec := podSpec(stack, volumeDefs, service, os)

	pts := v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      servicelabels.ServiceLabels(stack, service),
			Annotations: servicelabels.SafeMerge(nil, service.Spec.Metadata),
		},
		Spec: podSpec,
	}

	for _, volume := range podSpec.Volumes {
		config, ok := configsByName[strings.TrimPrefix(volume.Name, "config-")]
		if ok {
			h, err := config.Hash()
			if err == nil {
				pts.Annotations[fmt.Sprintf("rio.cattle.io/%s-hash", config.Name)] = h
			}
		}
	}

	return pts
}

func podSpec(stack *riov1.Stack, volumes map[string]*riov1.Volume, service *riov1.Service, os *objectset.ObjectSet) v1.PodSpec {
	var (
		f    = false
		spec = &service.Spec.ServiceUnversionedSpec
	)

	podSpec := v1.PodSpec{
		HostNetwork:                  spec.NetworkMode == "host",
		HostIPC:                      spec.IpcMode == "host",
		HostPID:                      spec.PidMode == "pid",
		Hostname:                     spec.Hostname,
		AutomountServiceAccountToken: &f,
	}

	podvolume.Populate(volumes, service, &podSpec)

	containers(&podSpec, service)
	dns(&podSpec, spec)
	restartPolicy(&podSpec, spec)
	stopPeriod(&podSpec, spec)
	scheduling(&podSpec, spec, servicelabels.ServiceLabels(stack, service))
	roles(stack, service, &podSpec, os)

	return podSpec
}

func roles(stack *riov1.Stack, service *riov1.Service, podSpec *v1.PodSpec, os *objectset.ObjectSet) {
	if err := rbac.Populate(stack, service, os); err != nil {
		os.AddErr(err)
		return
	}

	serviceAccountName := rbac.ServiceAccountName(service)
	if serviceAccountName != "" {
		podSpec.ServiceAccountName = serviceAccountName
		podSpec.AutomountServiceAccountToken = nil
	}
}

func stopPeriod(podSpec *v1.PodSpec, service *riov1.ServiceUnversionedSpec) {
	if service.StopGracePeriodSeconds != nil {
		v := int64(*service.StopGracePeriodSeconds)
		podSpec.TerminationGracePeriodSeconds = &v
	}
}

func restartPolicy(podSpec *v1.PodSpec, service *riov1.ServiceUnversionedSpec) {
	switch service.RestartPolicy {
	case "never":
		podSpec.RestartPolicy = v1.RestartPolicyNever
	case "on-failure":
		podSpec.RestartPolicy = v1.RestartPolicyOnFailure
	case "always":
		podSpec.RestartPolicy = v1.RestartPolicyAlways
	}
}

func containers(podSpec *v1.PodSpec, service *riov1.Service) {
	if service.Spec.Image != "" {
		podSpec.Containers = append(podSpec.Containers, container.Container(service.Name, service.Spec.ContainerConfig))
	}

	for _, name := range sidekick2.SortedKeys(service.Spec.Sidekicks) {
		sidekick := service.Spec.Sidekicks[name]
		c := container.Container(name, sidekick.ContainerConfig)
		if sidekick.InitContainer {
			podSpec.InitContainers = append(podSpec.InitContainers, c)
		} else {
			podSpec.Containers = append(podSpec.Containers, c)
		}
	}
}

func scheduling(podSpec *v1.PodSpec, service *riov1.ServiceUnversionedSpec, labels map[string]string) {
	nodeAffinity, err := service.Scheduling.ToNodeAffinity()
	if err == nil {
		podSpec.Affinity = &v1.Affinity{
			NodeAffinity: nodeAffinity,
		}
	} else {
		logrus.Errorf("failed to parse scheduling for service: %v", err)
	}

	podSpec.SchedulerName = service.Scheduling.Scheduler

	// mergeLabels will strip out rio.cattle.io labels
	for k, v := range servicelabels.SafeMerge(nil, labels) {
		toleration := v1.Toleration{
			Key:      k,
			Operator: v1.TolerationOpExists,
			Value:    v,
		}

		if len(toleration.Value) > 0 {
			toleration.Operator = v1.TolerationOpEqual
		}

		toleration.Effect = v1.TaintEffectNoExecute
		podSpec.Tolerations = append(podSpec.Tolerations)
		toleration.Effect = v1.TaintEffectNoSchedule
		podSpec.Tolerations = append(podSpec.Tolerations)
		toleration.Effect = v1.TaintEffectPreferNoSchedule
		podSpec.Tolerations = append(podSpec.Tolerations)
	}
}

func dns(podSpec *v1.PodSpec, service *riov1.ServiceUnversionedSpec) {
	dnsConfig := &v1.PodDNSConfig{
		Nameservers: service.DNS,
		Searches:    service.DNSSearch,
	}

	if len(dnsConfig.Nameservers) > 0 {
		podSpec.DNSPolicy = v1.DNSNone
	}

	var ns []string
	for _, name := range dnsConfig.Nameservers {
		if name == "host" {
			podSpec.DNSPolicy = v1.DNSDefault
		} else if name == "cluster" {
			if service.NetworkMode == "host" {
				podSpec.DNSPolicy = v1.DNSClusterFirstWithHostNet
			} else {
				podSpec.DNSPolicy = v1.DNSClusterFirst
			}
		} else {
			ns = append(ns, name)
		}
	}
	dnsConfig.Nameservers = ns

	for _, dnsOpt := range service.DNSOptions {
		k, v := kv.Split(dnsOpt, "=")
		opt := v1.PodDNSConfigOption{
			Name: k,
		}
		if len(v) > 0 {
			opt.Value = &v
		}
		dnsConfig.Options = append(dnsConfig.Options, opt)
	}

	if len(dnsConfig.Options) > 0 || len(dnsConfig.Searches) > 0 || len(dnsConfig.Nameservers) > 0 {
		podSpec.DNSConfig = dnsConfig
	}

	for _, host := range service.ExtraHosts {
		ip, host := kv.Split(host, ":")
		podSpec.HostAliases = append(podSpec.HostAliases, v1.HostAlias{
			IP:        ip,
			Hostnames: []string{host},
		})
	}
}
