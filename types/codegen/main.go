package main

import (
	"github.com/rancher/norman/generator"
	networkingSchema "github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3/schema"
	spaceSchema "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1/schema"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1/schema"
	"github.com/sirupsen/logrus"
)

var (
	basePackage = "github.com/rancher/rio/types"
)

func main() {
	if err := generator.DefaultGenerate(schema.Schemas, basePackage, true, nil); err != nil {
		logrus.Fatal(err)
	}
	if err := generator.DefaultGenerate(schema.Schemas, basePackage, true, nil); err != nil {
		logrus.Fatal(err)
	}
	if err := generator.DefaultGenerate(networkingSchema.Schemas, basePackage, false, nil); err != nil {
		logrus.Fatal(err)
	}
	if err := generator.DefaultGenerate(spaceSchema.Schemas, basePackage, true, nil); err != nil {
		logrus.Fatal(err)
	}
}
