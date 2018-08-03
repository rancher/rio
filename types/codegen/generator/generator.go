package generator

import (
	"path"
	"strings"

	"github.com/rancher/norman/generator"
	"github.com/rancher/norman/types"
)

var (
	basePackage = "github.com/rancher/rio/types"
	baseCattle  = "client"
	baseK8s     = "apis"
)

func Generate(schemas *types.Schemas, backendTypes map[string]bool) {
	version := getVersion(schemas)
	group := strings.Split(version.Group, ".")[0]

	cattleOutputPackage := path.Join(basePackage, baseCattle, group, version.Version)
	k8sOutputPackage := path.Join(basePackage, baseK8s, version.Group, version.Version)

	if err := generator.Generate(schemas, backendTypes, cattleOutputPackage, k8sOutputPackage); err != nil {
		panic(err)
	}
}

func getVersion(schemas *types.Schemas) *types.APIVersion {
	var version types.APIVersion
	for _, schema := range schemas.Schemas() {
		if version.Group == "" {
			version = schema.Version
			continue
		}
		if version.Group != schema.Version.Group ||
			version.Version != schema.Version.Version {
			panic("schema set contains two APIVersions")
		}
	}

	return &version
}
