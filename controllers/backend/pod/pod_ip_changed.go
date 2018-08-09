package pod

import "k8s.io/api/core/v1"

const (
	hostIPLabel = "rio.cattle.io/pod-ip"
)

func (p *Controller) checkChangedIP(pod *v1.Pod) (bool, error) {
	if pod.Status.PodIP == "" {
		return true, nil
	}

	ip := pod.Labels[hostIPLabel]
	if ip == "" {
		pod = pod.DeepCopy()
		pod.Labels[hostIPLabel] = pod.Status.PodIP
		_, err := p.pods.Update(pod)
		return false, err
	} else if ip != pod.Status.PodIP {
		err := p.pods.Delete(pod.Name, nil)
		return false, err
	}

	return true, nil
}
