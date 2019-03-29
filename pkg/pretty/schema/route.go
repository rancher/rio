package schema

import (
	"time"

	"github.com/rancher/mapper"
	m "github.com/rancher/mapper/mappers"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	pm "github.com/rancher/rio/pkg/pretty/mapper"
	"github.com/rancher/rio/pkg/pretty/objectmappers"
)

func route(schemas *mapper.Schemas) *mapper.Schemas {
	return schemas.
		AddMapperForType(v1.Fault{},
			pm.Duration{Field: "delayMillis", Unit: time.Millisecond},
			m.Move{From: "delayMillis", To: "delay"},
			m.AliasField{Field: "delay", Names: []string{"delayMillis"}},
			m.AliasField{Field: "percentage", Names: []string{"percent"}},
		).
		AddMapperForType(v1.Match{},
			pm.Destination{Field: "from"},
			pm.StringMatchMap{Field: "cookies"},
			pm.StringMatchMap{Field: "headers"},
			objectmappers.NewStringMatch("method"),
			objectmappers.NewStringMatch("scheme"),
			objectmappers.NewStringMatch("path"),
		).
		AddMapperForType(v1.Retry{},
			pm.Duration{Field: "timeoutMillis", Unit: time.Millisecond},
			m.Move{From: "timeoutMillis", To: "timeout"},
			m.AliasField{Field: "timeout", Names: []string{"timeoutMillis"}},
		).
		AddMapperForType(v1.RouteSpec{},
			m.MapToSlice{Field: "addHeaders", Sep: "="},
			objectmappers.NewMatch("matches"),
			m.SingleSlice{Field: "matches"},
			m.AliasField{Field: "matches", Names: []string{"match"}},
			pm.Destination{Field: "mirror"},
			pm.HostPath{Field: "redirect"},
			pm.HostPath{Field: "rewrite"},
			pm.Duration{Field: "timeoutMillis", Unit: time.Millisecond},
			m.Move{From: "timeoutMillis", To: "timeout"},
			m.AliasField{Field: "timeout", Names: []string{"timeoutMillis"}},
			pm.To{Field: "to"},
			m.SingleSlice{Field: "to"},
		)
}
