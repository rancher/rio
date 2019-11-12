package prometheus

import (
	"context"

	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	linkerdNamespace        = "linkerd"
	prometheusConfigMapName = "linkerd-prometheus-config"
	configKey               = "prometheus.yml"
	content                 = `global:
  scrape_interval: 10s
  scrape_timeout: 10s
  evaluation_interval: 10s
rule_files:
- /etc/prometheus/*_rules.yml
scrape_configs:
- job_name: 'prometheus'
  static_configs:
  - targets: ['localhost:9090']
- job_name: 'grafana'
  kubernetes_sd_configs:
  - role: pod
    namespaces:
      names: ['linkerd']
  relabel_configs:
  - source_labels:
    - __meta_kubernetes_pod_container_name
    action: keep
    regex: ^grafana$
#  Required for: https://grafana.com/grafana/dashboards/315
- job_name: 'kubernetes-nodes-cadvisor'
  scheme: https
  tls_config:
    ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
    insecure_skip_verify: true
  bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
  kubernetes_sd_configs:
  - role: node
  relabel_configs:
  - action: labelmap
    regex: __meta_kubernetes_node_label_(.+)
  - target_label: __address__
    replacement: kubernetes.default.svc:443
  - source_labels: [__meta_kubernetes_node_name]
    regex: (.+)
    target_label: __metrics_path__
    replacement: /api/v1/nodes/$1/proxy/metrics/cadvisor
- job_name: 'linkerd-controller'
  kubernetes_sd_configs:
  - role: pod
    namespaces:
      names: ['linkerd']
  relabel_configs:
  - source_labels:
    - __meta_kubernetes_pod_label_linkerd_io_control_plane_component
    - __meta_kubernetes_pod_container_port_name
    action: keep
    regex: (.*);admin-http$
  - source_labels: [__meta_kubernetes_pod_container_name]
    action: replace
    target_label: component
- job_name: 'linkerd-proxy'
  kubernetes_sd_configs:
  - role: pod
  relabel_configs:
  - source_labels:
    - __meta_kubernetes_pod_container_name
    - __meta_kubernetes_pod_container_port_name
    - __meta_kubernetes_pod_label_linkerd_io_control_plane_ns
    action: keep
    regex: ^linkerd-proxy;linkerd-admin;linkerd$
  - source_labels: [__meta_kubernetes_namespace]
    action: replace
    target_label: namespace
  - source_labels: [__meta_kubernetes_pod_name]
    action: replace
    target_label: pod
  # special case k8s' "job" label, to not interfere with prometheus' "job"
  # label
  # __meta_kubernetes_pod_label_linkerd_io_proxy_job=foo =>
  # k8s_job=foo
  - source_labels: [__meta_kubernetes_pod_label_linkerd_io_proxy_job]
    action: replace
    target_label: k8s_job
  # drop __meta_kubernetes_pod_label_linkerd_io_proxy_job
  - action: labeldrop
    regex: __meta_kubernetes_pod_label_linkerd_io_proxy_job
  # __meta_kubernetes_pod_label_linkerd_io_proxy_deployment=foo =>
  # deployment=foo
  - action: labelmap
    regex: __meta_kubernetes_pod_label_linkerd_io_proxy_(.+)
  # drop all labels that we just made copies of in the previous labelmap
  - action: labeldrop
    regex: __meta_kubernetes_pod_label_linkerd_io_proxy_(.+)
  # __meta_kubernetes_pod_label_linkerd_io_foo=bar =>
  # foo=bar
  - action: labelmap
    regex: __meta_kubernetes_pod_label_(.+)`
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := handler{
		configmaps: rContext.Core.Core().V1().ConfigMap(),
		pods:       rContext.Core.Core().V1().Pod(),
	}
	rContext.Core.Core().V1().ConfigMap().OnChange(ctx, "watch-proxy-injector", h.onChange)

	return nil
}

type handler struct {
	configmaps corev1controller.ConfigMapClient
	pods       corev1controller.PodClient
}

func (h handler) onChange(key string, obj *v1.ConfigMap) (*v1.ConfigMap, error) {
	if obj == nil || obj.DeletionTimestamp != nil {
		return obj, nil
	}

	if obj.Namespace == linkerdNamespace && obj.Name == prometheusConfigMapName {
		if obj.Data[configKey] == content {
			return obj, nil
		}

		dp := obj.DeepCopy()
		dp.Data[configKey] = content
		updated, err := h.configmaps.Update(dp)
		if err != nil {
			return obj, err
		}

		pods, err := h.pods.List(linkerdNamespace, metav1.ListOptions{
			LabelSelector: "linkerd.io/control-plane-component=prometheus",
		})
		if err != nil {
			return obj, err
		}
		for _, pod := range pods.Items {
			if err := h.pods.Delete(linkerdNamespace, pod.Name, &metav1.DeleteOptions{}); err != nil {
				return updated, err
			}
		}

		return updated, nil
	}

	return obj, nil

}
