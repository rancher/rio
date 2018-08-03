package mapper

import (
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/norman/types/values"
)

type DeploymentStatus struct {
}

func (s *DeploymentStatus) FromInternal(data map[string]interface{}) {
	setScaleStatuses(data)

	conditions := convert.ToMapSlice(values.GetValueN(data, "conditions"))
	conditions = append(conditions, convert.ToMapSlice(values.GetValueN(data, "deploymentStatus", "conditions"))...)
	if len(conditions) > 0 {
		values.PutValue(data, conditions, "conditions")
	}
}

func setScaleStatuses(data map[string]interface{}) {
	v, ok := values.GetValue(data, "deploymentStatus")
	if !ok {
		return
	}

	deploymentStatus := convert.ToMapInterface(v)
	if len(deploymentStatus) == 0 {
		return
	}

	scaleStatus := toScaleStatus(deploymentStatus)
	values.PutValue(data, scaleStatus, "scaleStatus")
}

func toScaleStatus(deploymentStatus map[string]interface{}) map[string]interface{} {
	ready, _ := convert.ToNumber(deploymentStatus["readyReplicas"])
	available, _ := convert.ToNumber(deploymentStatus["availableReplicas"])
	unavailable, _ := convert.ToNumber(deploymentStatus["unavailableReplicas"])
	updated, _ := convert.ToNumber(deploymentStatus["updatedReplicas"])

	return map[string]interface{}{
		"ready":       ready,
		"available":   available - ready,
		"unavailable": unavailable,
		"updated":     updated,
	}
}

func (s *DeploymentStatus) ToInternal(data map[string]interface{}) error {
	return nil
}

func (s *DeploymentStatus) ModifySchema(schema *types.Schema, schemas *types.Schemas) error {
	return nil
}
