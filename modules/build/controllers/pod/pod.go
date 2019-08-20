package pod

import (
	"context"

	adminv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	corev1 "k8s.io/api/core/v1"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := handler{
		systemNamespace: rContext.Namespace,
		infos:           rContext.Global.Admin().V1().RioInfo(),
	}
	rContext.Core.Core().V1().Pod().OnChange(ctx, "local-builder-pod-watcher", h.watchPod)

	return nil
}

type handler struct {
	systemNamespace string
	infos           adminv1controller.RioInfoController
}

func (h handler) watchPod(key string, pod *corev1.Pod) (*corev1.Pod, error) {
	if pod == nil || pod.Namespace != h.systemNamespace {
		return pod, nil
	}

	if pod.Labels["app"] == "buildkitd-dev" && pod.Status.Phase == corev1.PodRunning {
		info, err := h.infos.Cache().Get("rio")
		if err != nil {
			return pod, err
		}
		info.Status.BuildkitPodName = pod.Name
		if _, err := h.infos.Update(info); err != nil {
			return pod, err
		}
	}

	if pod.Labels["app"] == "socat-socket" && pod.Status.Phase == corev1.PodRunning {
		info, err := h.infos.Cache().Get("rio")
		if err != nil {
			return pod, err
		}
		info.Status.SocatPodName = pod.Name
		if _, err := h.infos.Update(info); err != nil {
			return pod, err
		}
	}

	return pod, nil
}
