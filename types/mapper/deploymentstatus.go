package mapper

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/norman/types/values"
)

var keys = []string{
	"deploymentStatus",
	"daemonSetStatus",
	"statefulStatus",
}

type DeploymentStatus struct {
}

func (s *DeploymentStatus) FromInternal(data map[string]interface{}) {
	setDeploymentScaleStatuses(data)
	setDaemonSetScaleStatuses(data)
	setStatefulSetScaleStatuses(data)

	conditions := convert.ToMapSlice(values.GetValueN(data, "conditions"))
	for _, key := range keys {
		conditions = append(conditions, convert.ToMapSlice(values.GetValueN(data, key, "conditions"))...)
	}
	if len(conditions) > 0 {
		values.PutValue(data, conditions, "conditions")
	}
}

func setDeploymentScaleStatuses(data map[string]interface{}) {
	v, ok := values.GetValue(data, "deploymentStatus")
	if !ok {
		return
	}

	deploymentStatus := convert.ToMapInterface(v)
	if len(deploymentStatus) == 0 {
		return
	}

	ready, _ := convert.ToNumber(deploymentStatus["readyReplicas"])
	available, _ := convert.ToNumber(deploymentStatus["availableReplicas"])
	unavailable, _ := convert.ToNumber(deploymentStatus["unavailableReplicas"])
	updated, _ := convert.ToNumber(deploymentStatus["updatedReplicas"])

	scaleStatus := map[string]interface{}{
		"ready":       ready,
		"available":   available - ready,
		"unavailable": unavailable,
		"updated":     updated,
	}

	values.PutValue(data, scaleStatus, "scaleStatus")
}

func setStatefulSetScaleStatuses(data map[string]interface{}) {
	v, ok := values.GetValue(data, "statefulSetStatus")
	if !ok {
		return
	}

	deploymentStatus := convert.ToMapInterface(v)
	if len(deploymentStatus) == 0 {
		return
	}

	ready, _ := convert.ToNumber(deploymentStatus["readyReplicas"])
	available, _ := convert.ToNumber(deploymentStatus["currentReplicas"])
	updated, _ := convert.ToNumber(deploymentStatus["updatedReplicas"])

	scaleStatus := map[string]interface{}{
		"ready":       ready,
		"available":   available - ready,
		"unavailable": 0,
		"updated":     updated,
	}

	values.PutValue(data, scaleStatus, "scaleStatus")
}

func setDaemonSetScaleStatuses(data map[string]interface{}) {
	v, ok := values.GetValue(data, "daemonSetStatus")
	if !ok {
		return
	}

	deploymentStatus := convert.ToMapInterface(v)
	if len(deploymentStatus) == 0 {
		return
	}

	ready, _ := convert.ToNumber(deploymentStatus["numberReady"])
	available, _ := convert.ToNumber(deploymentStatus["numberAvailable"])
	unavailable, _ := convert.ToNumber(deploymentStatus["numberUnavailable"])
	updated, _ := convert.ToNumber(deploymentStatus["updatedNumberScheduled"])
	scale, _ := convert.ToNumber(deploymentStatus["desiredNumberScheduled"])

	scaleStatus := map[string]interface{}{
		"ready":       ready,
		"available":   available - ready,
		"unavailable": unavailable,
		"updated":     updated,
	}

	values.PutValue(data, scaleStatus, "scaleStatus")
	values.PutValue(data, scale, "scale")
}

func (s *DeploymentStatus) ToInternal(data map[string]interface{}) error {
	return nil
}

func (s *DeploymentStatus) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return nil
}
