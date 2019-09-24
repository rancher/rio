package proxy

import (
	"context"
	"fmt"

	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	proxyImage = "rancher/klipper-lb:v0.1.2"
	socatImage = "alpine/socat:1.0.3"
	T          = true
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := Handler{
		namespace: rContext.Namespace,
		apply:     rContext.Apply.WithCacheTypes(rContext.Apps.Apps().V1().DaemonSet()).WithStrictCaching(),
	}

	rContext.Core.Core().V1().Service().OnChange(ctx, "registry-proxy", h.onChange)
	return nil
}

type Handler struct {
	namespace string
	apply     apply.Apply
}

func (h *Handler) onChange(key string, svc *corev1.Service) (*corev1.Service, error) {
	if svc == nil {
		return nil, nil
	}

	if svc.Namespace != h.namespace || svc.Name != "registry" {
		return svc, nil
	}

	if svc.Spec.ClusterIP == "" {
		return svc, nil
	}

	socatDeploy := constructors.NewDaemonset(svc.Namespace, "socat", v1.Deployment{
		ObjectMeta: v12.ObjectMeta{
			OwnerReferences: []v12.OwnerReference{
				{
					Name:       svc.Name,
					UID:        svc.UID,
					Kind:       "Service",
					APIVersion: "v1",
				},
			},
		},
		Spec: v1.DeploymentSpec{
			Selector: &v12.LabelSelector{
				MatchLabels: map[string]string{
					"app": "socat",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v12.ObjectMeta{
					Labels: map[string]string{
						"app": "socat",
					},
				},
				Spec: corev1.PodSpec{
					HostNetwork: true,
					Containers: []corev1.Container{
						{
							Name:  "socat",
							Image: socatImage,
							Command: []string{
								"socat",
								"TCP-LISTEN:5442,fork,bind=127.0.0.1,reuseaddr",
								fmt.Sprintf("TCP:%s:80", svc.Spec.ClusterIP),
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: &T,
							},
						},
					},
				},
			},
		},
	})

	os := objectset.NewObjectSet()
	os.Add(socatDeploy)
	return svc, h.apply.WithSetID(svc.Name + "-proxy").WithOwner(svc).Apply(os)
}
