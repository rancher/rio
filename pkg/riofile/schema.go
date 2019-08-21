package riofile

import (
	"github.com/rancher/mapper"
	"github.com/rancher/mapper/mappers"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/pretty/stringers"
	m "github.com/rancher/rio/pkg/riofile/mappers"
	corev1 "k8s.io/api/core/v1"
)

func init() {
	schema.DefaultPostMappers = func() []mapper.Mapper {
		return []mapper.Mapper{
			mappers.JSONKeys{},
		}
	}

	schema.
		Init(service).
		Init(config).
		Init(router).
		Init(externalservice).
		MustImport(riofile{})
}

func config(schemas *mapper.Schemas) *mapper.Schemas {
	schemas.AddMapperForType(corev1.ConfigMap{},
		m.NewObject("ConfigMap", "v1"),
		m.NewConfigMapMapper("data"))
	return schemas
}

func router(schemas *mapper.Schemas) *mapper.Schemas {
	schemas.AddMapperForType(v1.Router{},
		m.NewObject("Router", "rio.cattle.io/v1"))
	return schemas
}

func externalservice(schemas *mapper.Schemas) *mapper.Schemas {
	schemas.AddMapperForType(v1.ExternalService{},
		m.NewObject("ExternalService", "rio.cattle.io/v1"))
	return schemas
}

func service(schemas *mapper.Schemas) *mapper.Schemas {
	schemas.AddMapperForType(v1.Service{},
		m.NewObject("Service", "rio.cattle.io/v1"))
	schemas.AddMapperForType(v1.ServiceSpec{},
		m.Scale{},
		mappers.Drop{Field: "maxScale"},
		mappers.Drop{Field: "minScale"},

		// stringer
		stringers.NewPermissions("permissions"),
		stringers.NewPermissions("globalPermissions"),
		stringers.NewHostAlias("hostAliases"),
		stringers.NewDNSOptions("dnsOptions"),

		// enums
		m.NewFuzzy("dnsPolicy", "None", "Default", "ClusterFirst", "ClusterFirstWithHostNet"),

		// slices
		mappers.SingleSlice{Field: "dnsNameservers"},
		mappers.SingleSlice{Field: "dnsSearches"},
		mappers.SingleSlice{Field: "dnsOptions"},
		mappers.SingleSlice{Field: "args"},

		// aliases
		mappers.NewAlias("dnsNameservers", "dns", "nameservers", "nameserver"),
		mappers.NewAlias("dnsSearches", "dnsSearch", "search", "searches"),
		mappers.NewAlias("globalPermissions", "globalPermission"),
		mappers.NewAlias("hostAliases", "extraHosts", "addHosts", "hostAlias", "extraHost", "addHost"),
		mappers.NewAlias("permissions", "permission"),

		containerMappers(),
	)
	schemas.AddMapperForType(v1.NamedContainer{}, containerMappers())

	return schemas
}

func containerMappers() mapper.Mapper {
	return mapper.Mappers{
		// stringers
		stringers.NewContainerPort("ports"),
		stringers.NewEnv("env"),
		stringers.NewConfigs("configs"),
		stringers.NewSecrets("secrets"),
		stringers.NewVolume("volumes"),

		// parsers
		m.NewQuantity("cpus"),
		m.NewQuantity("memory"),

		// misc
		mappers.Shlex{Field: "command"},
		mappers.Shlex{Field: "args"},
		m.NewFuzzy("imagePullPolicy", "Always", "Never", "IfNotPresent"),

		// structures
		mappers.SingleSlice{Field: "configs", DontForceString: true},
		mappers.SingleSlice{Field: "secrets", DontForceString: true},
		mappers.SingleSlice{Field: "env", DontForceString: true},
		mappers.MapToSlice{Field: "env", Sep: "="},
		mappers.MapToSlice{Field: "configs", Sep: ":"},
		mappers.MapToSlice{Field: "secrets", Sep: ":"},

		m.HostNetwork{},

		mappers.NewAlias("args", "arg"),
		mappers.NewAlias("configs", "config"),
		mappers.NewAlias("cpus", "cpu"),
		mappers.NewAlias("env", "environment"),
		mappers.NewAlias("imagePullPolicy", "pullPolicy"),
		mappers.NewAlias("memory", "mem"),
		mappers.NewAlias("readOnlyRootFilesystem", "readOnly", "readOnlyFS"),
		mappers.NewAlias("runAsGroup", "group"),
		mappers.NewAlias("runAsUser", "user"),
		mappers.NewAlias("secrets", "secret"),
		mappers.NewAlias("stdin", "stdinOpen"),
	}
}
