package main

import (
	autoscalev1 "github.com/rancher/rio/pkg/apis/autoscale.rio.cattle.io/v1"
	projectv1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	controllergen "github.com/rancher/wrangler/pkg/controller-gen"
	"github.com/rancher/wrangler/pkg/controller-gen/args"
)

var (
	basePackage = "github.com/rancher/rio/types"
)

func main() {
	controllergen.Run(args.Options{
		OutputPackage: "github.com/rancher/rio/pkg/generated",
		Boilerplate:   "scripts/boilerplate.go.txt",
		Groups: map[string]args.Group{
			"project.rio.cattle.io": {
				Types: []interface{}{
					projectv1.ClusterDomain{},
					projectv1.Feature{},
				},
				GenerateTypes: true,
			},
			"autoscale.rio.cattle.io": {
				Types: []interface{}{
					autoscalev1.ServiceScaleRecommendation{},
				},
				GenerateTypes: true,
			},
			"rio.cattle.io": {
				Types: []interface{}{
					riov1.ExternalService{},
					riov1.Router{},
					riov1.Service{},
					riov1.PublicDomain{},
					riov1.App{},
				},
				GenerateTypes: true,
			},
		},
	})
}
