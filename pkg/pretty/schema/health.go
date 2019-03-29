package schema

import (
	"time"

	"github.com/rancher/mapper"
	m "github.com/rancher/mapper/mappers"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	pm "github.com/rancher/rio/pkg/pretty/mapper"
)

func health(schemas *mapper.Schemas) *mapper.Schemas {
	return schemas.
		AddMapperForType(v1.HealthConfig{},
			m.Shlex{Field: "test"},
			m.Move{From: "intervalSeconds", To: "interval"},
			pm.Duration{Field: "interval", Unit: time.Second},
			m.AliasField{Field: "interval", Names: []string{"period", "periodSeconds"}},
			m.Move{From: "timeoutSeconds", To: "timeout"},
			pm.Duration{Field: "timeout", Unit: time.Second},
			m.Move{From: "initialDelaySeconds", To: "initialDelay"},
			pm.Duration{Field: "initialDelay", Unit: time.Second},
			m.AliasField{Field: "initialDelay", Names: []string{"startPeriod"}},
			m.AliasField{Field: "healthyThreshold", Names: []string{"retries", "successThreshold"}},
			m.AliasField{Field: "unhealthyThreshold", Names: []string{"failureThreshold"}},
		)
}
