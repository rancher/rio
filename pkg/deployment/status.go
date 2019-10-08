package deployment

import (
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func IsReady(status *appv1.DeploymentStatus) bool {
	if status == nil {
		return false
	}

	for _, con := range status.Conditions {
		if con.Type == appv1.DeploymentAvailable && con.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}
