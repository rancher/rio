package main

import (
	splitv1alpha1 "github.com/deislabs/smi-sdk-go/pkg/apis/split/v1alpha1"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	v3 "github.com/rancher/rio/pkg/apis/management.cattle.io/v3"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	controllergen "github.com/rancher/wrangler/pkg/controller-gen"
	"github.com/rancher/wrangler/pkg/controller-gen/args"
	solov1 "github.com/solo-io/gloo/projects/gateway/pkg/api/v1/kube/apis/gateway.solo.io/v1"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1/kube/apis/gloo.solo.io/v1"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
)

var (
	basePackage = "github.com/rancher/rio/types"
)

func main() {
	controllergen.Run(args.Options{
		OutputPackage: "github.com/rancher/rio/pkg/generated",
		Boilerplate:   "scripts/boilerplate.go.txt",
		Groups: map[string]args.Group{
			"admin.rio.cattle.io": {
				Types: []interface{}{
					adminv1.ClusterDomain{},
					adminv1.RioInfo{},
					adminv1.PublicDomain{},
					adminv1.SystemStack{},
					adminv1.Certificate{},
				},
				GenerateTypes: true,
			},
			"rio.cattle.io": {
				Types: []interface{}{
					riov1.ExternalService{},
					riov1.Router{},
					riov1.Service{},
					riov1.Stack{},
				},
				GenerateTypes: true,
			},
			"gateway.solo.io": {
				Types: []interface{}{
					solov1.VirtualService{},
				},
				ClientSetPackage: "github.com/solo-io/gloo/projects/gateway/pkg/api/v1/kube/client/clientset/versioned",
				InformersPackage: "github.com/solo-io/gloo/projects/gateway/pkg/api/v1/kube/client/informers/externalversions",
				ListersPackage:   "github.com/solo-io/gloo/projects/gateway/pkg/api/v1/kube/client/listers",
			},
			"gloo.solo.io": {
				Types: []interface{}{
					gloov1.Settings{},
				},
				ClientSetPackage: "github.com/solo-io/gloo/projects/gloo/pkg/api/v1/kube/client/clientset/versioned",
				InformersPackage: "github.com/solo-io/gloo/projects/gloo/pkg/api/v1/kube/client/informers/externalversions",
				ListersPackage:   "github.com/solo-io/gloo/projects/gloo/pkg/api/v1/kube/client/listers",
			},
			"management.cattle.io": {
				Types: []interface{}{
					v3.Setting{},
					v3.User{},
				},
				GenerateTypes: true,
			},
			"split.smi-spec.io": {
				Types: []interface{}{
					splitv1alpha1.TrafficSplit{},
				},
				PackageName:     "split",
				GenerateClients: true,
			},
			"networking.istio.io": {
				Types: []interface{}{
					v1alpha3.Gateway{},
					v1alpha3.VirtualService{},
					v1alpha3.DestinationRule{},
					v1alpha3.ServiceEntry{},
				},
				PackageName:     "networking",
				GenerateClients: true,
			},
		},
	})
}
