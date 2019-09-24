package main

import (
	"os"

	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	autoscalev1 "github.com/rancher/rio/pkg/apis/autoscale.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	controllergen "github.com/rancher/wrangler/pkg/controller-gen"
	"github.com/rancher/wrangler/pkg/controller-gen/args"
)

var (
	basePackage = "github.com/rancher/rio/types"
)

func main() {
	os.Unsetenv("GOPATH")
	controllergen.Run(args.Options{
		OutputPackage: "github.com/rancher/rio/pkg/generated",
		Boilerplate:   "scripts/boilerplate.go.txt",
		Groups: map[string]args.Group{
			"admin.rio.cattle.io": {
				Types: []interface{}{
					adminv1.ClusterDomain{},
					adminv1.Feature{},
					adminv1.RioInfo{},
					adminv1.PublicDomain{},
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
					riov1.Stack{},
					riov1.ImageBuild{},
				},
				GenerateTypes: true,
			},
		},
	})
}
