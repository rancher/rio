package schema

import (
	"time"

	"github.com/rancher/mapper"
	m "github.com/rancher/mapper/mappers"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	pm "github.com/rancher/rio/pkg/pretty/mapper"
	"github.com/rancher/rio/pkg/pretty/objectmappers"
)

func containerMappers() []mapper.Mapper {
	return []mapper.Mapper{
		m.SingleSlice{Field: "capAdd"},
		m.SingleSlice{Field: "capDrop"},
		m.Shlex{Field: "command"},
		objectmappers.NewConfigMapping("configs"),
		m.MapToSlice{Field: "configs", Sep: ":"},
		m.SingleSlice{Field: "configs"},
		objectmappers.NewDeviceMapping("devices"),
		m.MapToSlice{Field: "devices", Sep: ":"},
		m.AliasField{Field: "environment", Names: []string{"env"}},
		m.MapToSlice{Field: "environment", Sep: "="},
		objectmappers.NewExposedPorts("expose"),
		pm.HealthMapper{Field: "healthcheck"},
		m.AliasField{Field: "imagePullPolicy", Names: []string{"pullPolicy"}},
		m.AliasValue{Field: "imagePullPolicy", Alias: map[string][]string{
			"always":      {"Always"},
			"never":       {"Never"},
			"not-present": {"IfNotPresent"}},
		},
		m.Move{From: "memoryLimitBytes", To: "memoryLimit"},
		m.Bytes{Field: "memoryLimit"},
		m.AliasField{Field: "memoryLimit", Names: []string{"memoryLimitsBytes"}},
		m.Move{From: "memoryReservationBytes", To: "memory"},
		m.Bytes{Field: "memory"},
		m.AliasField{Field: "memory", Names: []string{"mem", "memoryReservationBytes"}},
		m.Move{From: "nanoCpus", To: "cpus"},
		m.AliasField{Field: "cpus", Names: []string{"nanoCpus"}},
		objectmappers.NewPortBinding("ports"),
		pm.HealthMapper{Field: "readycheck"},
		objectmappers.NewSecretMapping("secrets"),
		m.MapToSlice{Field: "secrets", Sep: ":"},
		m.SingleSlice{Field: "secrets"},
		m.AliasField{Field: "stdinOpen", Names: []string{"interactive"}},
		objectmappers.NewTmpfs("tmpfs"),
		m.SingleSlice{Field: "tmpfs"},
		m.SingleSlice{Field: "volumesFrom"},
		objectmappers.NewMounts("volumes"),
		m.SingleSlice{Field: "volumes"},
	}
}

func serviceMappers() []mapper.Mapper {
	return append(containerMappers(),
		// Sorted by field name (mostly)
		m.SingleSlice{Field: "dns"},
		m.SingleSlice{Field: "dnsOptions"},
		m.SingleSlice{Field: "dnsSearch"},
		m.MapToSlice{Field: "extraHosts", Sep: ":"},
		m.AliasField{Field: "globalPermissions", Names: []string{"globalPerms"}},
		objectmappers.NewPermission("globalPermissions"),
		m.AliasField{Field: "metadata", Names: []string{"annotations"}},
		m.AliasField{Field: "net", Names: []string{"network"}},
		m.AliasValue{Field: "net", Alias: map[string][]string{
			"default": {"bridge"}},
		},
		m.AliasField{Field: "permissions", Names: []string{"perms"}},
		objectmappers.NewPermission("permissions"),
		m.AliasValue{Field: "restart", Alias: map[string][]string{
			"never":      {"no"},
			"on-failure": {"OnFailure"}},
		},
		m.DefaultMissing{Field: "scale", Default: 1},
		m.Move{From: "scheduling/node/nodeName", To: "node"},
		pm.SchedulingMapper{Field: "scheduling"},
		pm.Duration{Field: "stopGracePeriod", Unit: time.Second},
	)
}

func services(schemas *mapper.Schemas) *mapper.Schemas {
	return schemas.
		AddMapperForType(riov1.SidekickConfig{}, containerMappers()...).
		AddMapperForType(riov1.ServiceSpec{}, serviceMappers()...)
}
