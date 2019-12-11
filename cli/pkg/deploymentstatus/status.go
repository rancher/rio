package deploymentstatus

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func IsReady(status appsv1.DeploymentStatus) bool {
	for _, con := range status.Conditions {
		if con.Type == appsv1.DeploymentAvailable && con.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
