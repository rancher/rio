package service

import (
	"context"
	"fmt"
	"strings"

	approuter "github.com/rancher/rdns-server/client"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/config"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/generic"
	"github.com/rancher/wrangler/pkg/relatedresource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	rDNSClient := approuter.NewClient(rContext.Core.Core().V1().Secret(),
		rContext.Core.Core().V1().Secret().Cache(),
		rContext.Namespace)
	rDNSClient.SetBaseURL(constants.RDNSURL)

	h := handler{
		ctx:             ctx,
		systemNamespace: rContext.Namespace,
		configKey:       fmt.Sprintf("%s/%s", rContext.Namespace, config.ConfigName),
		configMapCache:  rContext.Core.Core().V1().ConfigMap().Cache(),
		endpointCache:   rContext.Core.Core().V1().Endpoints().Cache(),
		podCache:        rContext.Core.Core().V1().Pod().Cache(),
		nodeCache:       rContext.Core.Core().V1().Node().Cache(),
		services:        rContext.Core.Core().V1().Service(),
		rDNSClient:      rDNSClient,
	}

	corev1controller.RegisterServiceGeneratingHandler(ctx,
		rContext.Core.Core().V1().Service(),
		rContext.Apply.WithCacheTypes(rContext.Admin.Admin().V1().ClusterDomain()),
		"",
		"rdns-service",
		h.generate,
		&generic.GeneratingHandlerOptions{
			AllowClusterScoped: true,
		})

	rContext.Core.Core().V1().ConfigMap().OnChange(ctx, "rdns-config", h.onConfigMapChange)

	relatedresource.Watch(ctx, "rdns",
		func(namespace, name string, obj runtime.Object) ([]relatedresource.Key, error) {
			if name == h.gatewayName && namespace == h.gatewayNamespace {
				return []relatedresource.Key{
					{
						Namespace: namespace,
						Name:      name,
					},
				}, nil
			}
			return nil, nil
		},
		rContext.Core.Core().V1().Service(),
		rContext.Core.Core().V1().Endpoints())

	return nil
}

type handler struct {
	ctx context.Context

	started          bool
	gatewayName      string
	gatewayNamespace string
	systemNamespace  string
	configKey        string
	configMapCache   corev1controller.ConfigMapCache
	endpointCache    corev1controller.EndpointsCache
	podCache         corev1controller.PodCache
	nodeCache        corev1controller.NodeCache
	services         corev1controller.ServiceController
	rDNSClient       *approuter.Client
}

func (h *handler) onConfigMapChange(key string, cm *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	if cm == nil || h.configKey != key {
		return cm, nil
	}

	config, err := config.FromConfigMap(cm)
	if err != nil {
		return cm, err
	}

	if config.Gateway.ServiceName != h.gatewayName || config.Gateway.ServiceNamespace != h.gatewayNamespace {
		h.gatewayName = config.Gateway.ServiceName
		h.gatewayNamespace = config.Gateway.ServiceNamespace
		h.start()
	}

	if h.gatewayName != "" && h.gatewayNamespace != "" {
		h.services.Enqueue(h.gatewayNamespace, h.gatewayName)
	}

	return cm, nil
}

func (h *handler) generate(svc *corev1.Service, status corev1.ServiceStatus) ([]runtime.Object, corev1.ServiceStatus, error) {
	if h.gatewayName == "" || h.gatewayNamespace == "" || svc.Namespace != h.gatewayNamespace || svc.Name != h.gatewayName {
		return nil, status, generic.ErrSkip
	}

	addresses, nodePort, err := h.getAddresses(svc)
	if err != nil {
		return nil, status, err
	}

	if len(addresses) == 0 {
		return nil, status, nil
	}

	domainName, err := h.getDomain(addresses)
	if err != nil {
		return nil, status, err
	}

	if domainName == "" {
		return nil, status, nil
	}

	clusterDomain := &adminv1.ClusterDomain{
		ObjectMeta: metav1.ObjectMeta{
			Name:      domainName,
			Namespace: svc.Namespace,
		},
		Spec: adminv1.ClusterDomainSpec{
			Addresses: addresses,
		},
	}

	for _, port := range svc.Spec.Ports {
		portNum := 0
		if nodePort {
			portNum = int(port.NodePort)
		} else {
			portNum = int(port.Port)
		}

		if port.Name == "http" {
			clusterDomain.Spec.HTTPPort = portNum
		} else if port.Name == "https" {
			clusterDomain.Spec.HTTPSPort = portNum
		}
	}

	return []runtime.Object{
		clusterDomain,
	}, status, nil
}

func (h *handler) staticAddress() ([]adminv1.Address, error) {
	if config.ConfigController.IPAddresses != "" {
		ips := strings.Split(config.ConfigController.IPAddresses, ",")
		var addresses []adminv1.Address
		for _, ip := range ips {
			addresses = append(addresses, adminv1.Address{
				IP:       ip,
				Hostname: "",
			})
		}
		return addresses, nil
	}
	cm, err := h.configMapCache.Get(h.systemNamespace, config.ConfigName)
	if err != nil {
		return nil, err
	}

	config, err := config.FromConfigMap(cm)
	if err != nil {
		return nil, err
	}

	return config.Gateway.StaticAddresses, nil
}

func (h *handler) getAddresses(svc *corev1.Service) ([]adminv1.Address, bool, error) {
	result, err := h.staticAddress()
	if len(result) > 0 || err != nil {
		useNodePort := true
		for _, ingress := range svc.Status.LoadBalancer.Ingress {
			if ingress.IP != "" || ingress.Hostname != "" {
				useNodePort = false
			}
		}
		return result, useNodePort, err
	}

	if svc.Spec.Type == corev1.ServiceTypeExternalName {
		if svc.Spec.ExternalName == "" {
			return nil, false, nil
		}
		return []adminv1.Address{
			{
				Hostname: svc.Spec.ExternalName,
			},
		}, false, nil
	} else if svc.Spec.ClusterIP == corev1.ClusterIPNone {
		return h.getAddressHeadless(svc)
	}

	for _, ingress := range svc.Status.LoadBalancer.Ingress {
		if ingress.IP != "" || ingress.Hostname != "" {
			result = append(result, adminv1.Address{
				IP:       ingress.IP,
				Hostname: ingress.Hostname,
			})
		}
	}

	if len(result) > 0 {
		return result, false, nil
	}

	return h.getAddressNodePort(svc)
}

func (h *handler) getAddressHeadless(svc *corev1.Service) ([]adminv1.Address, bool, error) {
	endpoints, err := h.endpointCache.Get(svc.Namespace, svc.Name)
	if errors.IsNotFound(err) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}

	var result []adminv1.Address

	for _, endpoint := range endpoints.Subsets {
		for _, addr := range endpoint.Addresses {
			if addr.IP != "" {
				result = append(result, adminv1.Address{
					IP: addr.IP,
				})
			}
			if addr.Hostname != "" {
				result = append(result, adminv1.Address{
					Hostname: addr.Hostname,
				})
			}
		}
	}

	return result, false, nil
}

func (h *handler) getAddressNodePort(svc *corev1.Service) ([]adminv1.Address, bool, error) {
	endpoints, err := h.endpointCache.Get(svc.Namespace, svc.Name)
	if errors.IsNotFound(err) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}

	ips := map[string]bool{}

	for _, endpoint := range endpoints.Subsets {
		for _, addr := range endpoint.Addresses {
			if addr.TargetRef == nil {
				continue
			}
			if addr.TargetRef.Kind != "Pod" {
				continue
			}

			addr, err := h.getNodeIPForPod(addr.TargetRef.Namespace, addr.TargetRef.Name)
			if err != nil {
				return nil, false, err
			}

			if addr == "" {
				continue
			}

			ips[addr] = true
		}
	}

	var result []adminv1.Address
	for ip := range ips {
		result = append(result, adminv1.Address{
			IP: ip,
		})
	}

	return result, true, nil
}

func (h *handler) getNodeIPForPod(namespace, name string) (string, error) {
	pod, err := h.podCache.Get(namespace, name)
	if errors.IsNotFound(err) {
		return "", nil
	} else if err != nil {
		return "", err
	}

	node, err := h.nodeCache.Get(pod.Spec.NodeName)
	if errors.IsNotFound(err) {
		return "", nil
	} else if err != nil {
		return "", err
	}

	var (
		internal = ""
	)

	for _, nodeAddress := range node.Status.Addresses {
		switch nodeAddress.Type {
		case corev1.NodeExternalIP:
			return nodeAddress.Address, nil
		case corev1.NodeInternalIP:
			internal = nodeAddress.Address
		}
	}

	return internal, nil
}
