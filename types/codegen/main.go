package main

import (
	networkingSchema "github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3/schema"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1/schema"
	spaceSchema "github.com/rancher/rio/types/apis/space.cattle.io/v1beta1/schema"
	"github.com/rancher/rio/types/codegen/generator"
)

func main() {
	generator.Generate(schema.Schemas, nil)
	generator.Generate(networkingSchema.Schemas, map[string]bool{
		"virtualService": true,
		"gateway":        true,
	})
	generator.Generate(spaceSchema.Schemas, nil)
}
