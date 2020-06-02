package vsfactory

import (
	"time"

	rioadminv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	corev1 "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	extensionsv1beta1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/extensions/v1beta1"
)

const (
	rioNameHeader         = "X-Rio-ServiceName"
	rioNamespaceHeader    = "X-Rio-Namespace"
	scaleFromZeroAttempts = 10
)

// Retry configuration for the envoy gateway (Gloo)  https://www.envoyproxy.io/learn/automatic-retries
// for services scaled to zero
var overallTimeout = time.Second * 30
var perTryTimeout = time.Second * 2

type VirtualServiceFactory struct {
	secretCache        corev1controller.SecretCache
	clusterDomainCache rioadminv1controller.ClusterDomainCache
	publicDomainCache  rioadminv1controller.PublicDomainCache
	ingresses          extensionsv1beta1controller.IngressCache
	endpoints          corev1.EndpointsCache
	systemNamespace    string
}

func New(rContext *types.Context) *VirtualServiceFactory {
	return &VirtualServiceFactory{
		secretCache:        rContext.Core.Core().V1().Secret().Cache(),
		clusterDomainCache: rContext.Admin.Admin().V1().ClusterDomain().Cache(),
		publicDomainCache:  rContext.Admin.Admin().V1().PublicDomain().Cache(),
		systemNamespace:    rContext.Namespace,
		ingresses:          rContext.K8sNetworking.Extensions().V1beta1().Ingress().Cache(),
		endpoints:          rContext.Core.Core().V1().Endpoints().Cache(),
	}
}
