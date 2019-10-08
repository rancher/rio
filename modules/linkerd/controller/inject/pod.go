package inject

import (
	"context"

	"github.com/rancher/rio/pkg/deployment"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	name      = "linkerd-proxy-injector"
	namespace = "linkerd"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := handler{
		pods:      rContext.Core.Core().V1().Pod(),
		namespace: rContext.Namespace,
	}
	rContext.Apps.Apps().V1().Deployment().OnChange(ctx, "watch-proxy-injector", h.onChange)

	return nil
}

type handler struct {
	pods      corev1controller.PodController
	namespace string
}

func (h handler) onChange(key string, obj *v1.Deployment) (*v1.Deployment, error) {
	if obj == nil || obj.DeletionTimestamp != nil {
		return nil, nil
	}

	if obj.Namespace != namespace && obj.Name != name {
		return nil, nil
	}

	if !deployment.IsReady(&obj.Status) {
		return nil, nil
	}

	pods, err := h.pods.Cache().List(h.namespace, labels.NewSelector())
	if err != nil {
		return nil, err
	}
	for _, pod := range pods {
		if pod.Annotations["linkerd.io/inject"] == "enabled" {
			found := false
			for _, con := range pod.Spec.Containers {
				if con.Name == "linkerd-proxy" {
					found = true
					break
				}
			}
			if !found {
				if err := h.pods.Delete(pod.Namespace, pod.Name, &metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
					return nil, err
				}
			}
		}
	}
	return obj, nil
}
