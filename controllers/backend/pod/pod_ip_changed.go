package pod

import (
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
)

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
	} else if ip != pod.Status.PodIP || ip != pod.Status.HostIP {
		logrus.Infof("Deleting gateway %s because pod IP changed %s=>(%s/%s)", pod.Name, ip, pod.Status.PodIP,
			pod.Status.HostIP)
		err := p.pods.Delete(pod.Name, nil)
		return false, err
	}

	return true, nil
}
