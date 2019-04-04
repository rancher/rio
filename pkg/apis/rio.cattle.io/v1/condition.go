package v1

import (
	"github.com/rancher/wrangler/pkg/condition"
)

var (
	PendingCondition  = condition.Cond("Pending")
	DeployedCondition = condition.Cond("Deployed")
)
