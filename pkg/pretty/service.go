package pretty

import (
	"time"

	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/mapper"
	pm "github.com/rancher/rio/pkg/pretty/mapper"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

func containerMappers() []types.Mapper {
	return []types.Mapper{
		pm.SingleSlice{Field: "capAdd"},
		pm.SingleSlice{Field: "capDrop"},
		pm.Shlex{Field: "command"},
		pm.NewConfigMapping("configs"),
		pm.MapToSlice{Field: "configs", Sep: ":"},
		pm.SingleSlice{Field: "configs"},
		pm.NewDeviceMapping("devices"),
		pm.MapToSlice{Field: "devices", Sep: ":"},
		pm.AliasField{Field: "environment", Names: []string{"env"}},
		pm.MapToSlice{Field: "environment", Sep: "="},
		pm.NewExposedPorts("expose"),
		pm.HealthMapper{Field: "healthcheck"},
		pm.AliasField{Field: "imagePullPolicy", Names: []string{"pullPolicy"}},
		pm.AliasValue{Field: "imagePullPolicy", Alias: map[string][]string{
			"always":      {"Always"},
			"never":       {"Never"},
			"not-present": {"IfNotPresent"}},
		},
		mapper.Move{From: "memoryLimitBytes", To: "memoryLimit"},
		pm.Bytes{Field: "memoryLimit"},
		pm.AliasField{Field: "memoryLimit", Names: []string{"memoryLimitsBytes"}},
		mapper.Move{From: "memoryReservationBytes", To: "memory"},
		pm.Bytes{Field: "memory"},
		pm.AliasField{Field: "memory", Names: []string{"mem", "memoryReservationBytes"}},
		mapper.Move{From: "nanoCpus", To: "cpus"},
		pm.AliasField{Field: "cpus", Names: []string{"nanoCpus"}},
		pm.NewSecretMapping("secrets"),
		pm.MapToSlice{Field: "secrets", Sep: ":"},
		pm.SingleSlice{Field: "secrets"},
		pm.AliasField{Field: "stdinOpen", Names: []string{"interactive"}},
		pm.NewTmpfs("tmpfs"),
		pm.SingleSlice{Field: "tmpfs"},
		pm.SingleSlice{Field: "volumesFrom"},
		pm.NewMounts("volumes"),
		pm.SingleSlice{Field: "volumes"},
	}
}

func serviceMappers() []types.Mapper {
	return append(containerMappers(),
		// Sorted by field name (mostly)
		pm.SingleSlice{Field: "dns"},
		pm.SingleSlice{Field: "dnsOptions"},
		pm.SingleSlice{Field: "dnsSearch"},
		pm.MapToSlice{Field: "extraHosts", Sep: ":"},
		pm.AliasField{Field: "globalPermissions", Names: []string{"globalPerms"}},
		pm.NewPermission("globalPermissions"),
		pm.AliasField{Field: "metadata", Names: []string{"annotations"}},
		pm.AliasField{Field: "net", Names: []string{"network"}},
		pm.AliasValue{Field: "net", Alias: map[string][]string{
			"default": {"bridge"}},
		},
		pm.AliasField{Field: "permissions", Names: []string{"perms"}},
		pm.NewPermission("permissions"),
		pm.NewPortBinding("ports"),
		pm.AliasValue{Field: "restart", Alias: map[string][]string{
			"never":      {"no"},
			"on-failure": {"OnFailure"}},
		},
		pm.DefaultMissing{Field: "scale", Default: 1},
		mapper.Move{From: "scheduling/node/nodeId", To: "node"},
		pm.SchedulingMapper{Field: "scheduling"},
		mapper.Drop{Field: "spaceId", IgnoreDefinition: true},
		mapper.Drop{Field: "stackId", IgnoreDefinition: true},
		pm.Duration{Field: "stopGracePeriod", Unit: time.Second},
	)
}

func services(schemas *types.Schemas) *types.Schemas {
	return schemas.
		AddMapperForType(&Version, client.SidekickConfig{}, containerMappers()...).
		AddMapperForType(&Version, client.ServiceRevision{}, serviceMappers()...).
		AddMapperForType(&Version, client.Service{}, serviceMappers()...)
}
