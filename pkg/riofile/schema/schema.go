package schema

import (
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	m "github.com/rancher/rio/pkg/riofile/mappers"
	"github.com/rancher/rio/pkg/riofile/stringers"
	types "github.com/rancher/wrangler/pkg/schemas"
	mapper "github.com/rancher/wrangler/pkg/schemas/mappers"
	corev1 "k8s.io/api/core/v1"
)

var (
	Schema = types.EmptySchemas()
)

func init() {
	Schema.DefaultPostMapper = func() types.Mapper {
		return mapper.JSONKeys{}
	}

	Schema.
		Init(mappers).
		Init(service).
		Init(config).
		Init(router).
		Init(externalservice).
		TypeName("Riofile", ExternalRiofile{}).
		MustImport(ExternalRiofile{})
}

func mappers(schemas *types.Schemas) *types.Schemas {
	return objectToSlice(schemas).
		AddFieldMapper("alias", mapper.NewAlias).
		AddFieldMapper("duration", m.NewDuration).
		AddFieldMapper("quantity", m.NewQuantity).
		AddFieldMapper("enum", mapper.NewEnum).
		AddFieldMapper("hostNetwork", m.NewHostNetwork).
		AddFieldMapper("envmap", m.NewEnvMap).
		AddFieldMapper("shlex", m.NewShlex)
}

func objectToSlice(schemas *types.Schemas) *types.Schemas {
	schemas.AddFieldMapper("configs", m.NewObjectsToSliceFactory(
		func() m.MaybeStringer {
			return &stringers.ConfigsStringer{}
		},
		func(str string) (interface{}, error) {
			return stringers.ParseConfig(str)
		}))
	schemas.AddFieldMapper("secrets", m.NewObjectsToSliceFactory(
		func() m.MaybeStringer {
			return &stringers.SecretsStringer{}
		},
		func(str string) (interface{}, error) {
			return stringers.ParseSecret(str)
		}))
	schemas.AddFieldMapper("dnsOptions", m.NewObjectsToSliceFactory(
		func() m.MaybeStringer {
			return &stringers.PodDNSConfigOptionStringer{}
		},
		func(str string) (interface{}, error) {
			return stringers.ParseDNSOptions(str)
		}))
	schemas.AddFieldMapper("env", m.NewObjectsToSliceFactory(
		func() m.MaybeStringer {
			return &stringers.EnvStringer{}
		},
		func(str string) (interface{}, error) {
			return stringers.ParseEnv(str)
		}))
	schemas.AddFieldMapper("ports", m.NewObjectsToSliceFactory(
		func() m.MaybeStringer {
			return &stringers.ContainerPortStringer{}
		},
		func(str string) (interface{}, error) {
			return stringers.ParsePort(str)
		}))
	schemas.AddFieldMapper("hosts", m.NewObjectsToSliceFactory(
		func() m.MaybeStringer {
			return &stringers.HostAliasStringer{}
		},
		func(str string) (interface{}, error) {
			return stringers.ParseHostAlias(str)
		}))
	schemas.AddFieldMapper("volumes", m.NewObjectsToSliceFactory(
		func() m.MaybeStringer {
			return &stringers.VolumeStringer{}
		},
		func(str string) (interface{}, error) {
			return stringers.ParseVolume(str)
		}))
	schemas.AddFieldMapper("permissions", m.NewObjectsToSliceFactory(
		func() m.MaybeStringer {
			return &stringers.PermissionStringer{}
		},
		func(str string) (interface{}, error) {
			return stringers.ParsePermission(str)
		}))

	return schemas
}

func config(schemas *types.Schemas) *types.Schemas {
	schemas.AddMapperForType(corev1.ConfigMap{},
		m.NewObject("ConfigMap", "v1"),
		m.NewConfigMapMapper("data"))
	return schemas
}

func router(schemas *types.Schemas) *types.Schemas {
	schemas.AddMapperForType(v1.Router{},
		m.NewObject("Router", "rio.cattle.io/v1"))
	return schemas
}

func externalservice(schemas *types.Schemas) *types.Schemas {
	schemas.AddMapperForType(v1.ExternalService{},
		m.NewObject("ExternalService", "rio.cattle.io/v1"))
	return schemas
}

func service(schemas *types.Schemas) *types.Schemas {
	schemas.AddMapperForType(v1.Service{},
		m.NewObject("Service", "rio.cattle.io/v1"))

	return schemas
}
