package pretty

import (
	"time"

	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/mapper"
	pm "github.com/rancher/rio/pkg/pretty/mapper"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

func route(schemas *types.Schemas) *types.Schemas {
	return schemas.
		AddMapperForType(&Version, riov1.Fault{},
			pm.Duration{Field: "delayMillis", Unit: time.Millisecond},
			mapper.Move{From: "delayMillis", To: "delay"},
			pm.AliasField{Field: "delay", Names: []string{"delayMillis"}},
			pm.AliasField{Field: "percentage", Names: []string{"percent"}},
		).
		AddMapperForType(&Version, riov1.Match{},
			pm.Destination{Field: "from"},
			pm.StringMatchMap{Field: "cookies"},
			pm.StringMatchMap{Field: "headers"},
			pm.StringMatch{Field: "method"},
			pm.StringMatch{Field: "scheme"},
			pm.StringMatch{Field: "path"},
		).
		AddMapperForType(&Version, riov1.Retry{},
			pm.Duration{Field: "timeoutMillis", Unit: time.Millisecond},
			mapper.Move{From: "timeoutMillis", To: "timeout"},
			pm.AliasField{Field: "timeout", Names: []string{"timeoutMillis"}},
		).
		AddMapperForType(&Version, riov1.RouteSpec{},
			pm.MapToSlice{Field: "addHeaders", Sep: "="},
			pm.NewMatch("matches"),
			pm.SingleSlice{Field: "matches"},
			pm.AliasField{Field: "matches", Names: []string{"match"}},
			pm.Destination{Field: "mirror"},
			pm.HostPath{Field: "redirect"},
			pm.HostPath{Field: "rewrite"},
			pm.Duration{Field: "timeoutMillis", Unit: time.Millisecond},
			mapper.Move{From: "timeoutMillis", To: "timeout"},
			pm.AliasField{Field: "timeout", Names: []string{"timeoutMillis"}},
			pm.To{Field: "to"},
			pm.SingleSlice{Field: "to"},
		)
}
