package main

import (
	certmanagerv1alpha1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	"github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/knative/pkg/apis/istio/v1alpha3"
	autoscalev1 "github.com/rancher/rio/pkg/apis/autoscale.rio.cattle.io/v1"
	gitv1 "github.com/rancher/rio/pkg/apis/git.rio.cattle.io/v1"
	projectv1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	webhookv1 "github.com/rancher/rio/pkg/apis/webhookinator.rio.cattle.io/v1"
	controllergen "github.com/rancher/wrangler/pkg/controller-gen"
	"github.com/rancher/wrangler/pkg/controller-gen/args"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	extentionv1beta1 "k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

var (
	basePackage = "github.com/rancher/rio/types"
)

func main() {
	controllergen.Run(args.Options{
		OutputPackage: "github.com/rancher/rio/pkg/generated",
		Boilerplate:   "scripts/boilerplate.go.txt",
		Groups: map[string]args.Group{
			"git.rio.cattle.io": {
				Types: []interface{}{
					gitv1.GitModule{},
				},
				GenerateTypes: true,
			},
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
			"webhookinator.rio.cattle.io": {
				Types: []interface{}{
					webhookv1.GitWebHookReceiver{},
					webhookv1.GitWebHookExecution{},
				},
				GenerateTypes: true,
			},
			"": {
				Types: []interface{}{
					v1.Node{},
					v1.Namespace{},
					v1.Secret{},
					v1.Service{},
					v1.ServiceAccount{},
					v1.Endpoints{},
					v1.ConfigMap{},
					v1.PersistentVolumeClaim{},
					v1.Pod{},
				},
				InformersPackage: "k8s.io/client-go/informers",
				ClientSetPackage: "k8s.io/client-go/kubernetes",
				ListersPackage:   "k8s.io/client-go/listers",
			},
			"extensions": {
				Types: []interface{}{
					extentionv1beta1.Ingress{},
				},
				InformersPackage: "k8s.io/client-go/informers",
				ClientSetPackage: "k8s.io/client-go/kubernetes",
				ListersPackage:   "k8s.io/client-go/listers",
			},
			"rbac": {
				Types: []interface{}{
					rbacv1.Role{},
					rbacv1.RoleBinding{},
					rbacv1.ClusterRole{},
					rbacv1.ClusterRoleBinding{},
				},
				InformersPackage: "k8s.io/client-go/informers",
				ClientSetPackage: "k8s.io/client-go/kubernetes",
				ListersPackage:   "k8s.io/client-go/listers",
			},
			"apps": {
				Types: []interface{}{
					appsv1.Deployment{},
				},
				InformersPackage: "k8s.io/client-go/informers",
				ClientSetPackage: "k8s.io/client-go/kubernetes",
				ListersPackage:   "k8s.io/client-go/listers",
			},
			"storage": {
				Types: []interface{}{
					storagev1.StorageClass{},
				},
				InformersPackage: "k8s.io/client-go/informers",
				ClientSetPackage: "k8s.io/client-go/kubernetes",
				ListersPackage:   "k8s.io/client-go/listers",
			},
			"certmanager.k8s.io": {
				Types: []interface{}{
					certmanagerv1alpha1.Certificate{},
					certmanagerv1alpha1.ClusterIssuer{},
				},
				PackageName:      "certmanager",
				ClientSetPackage: "github.com/jetstack/cert-manager/pkg/client/clientset/versioned",
				InformersPackage: "github.com/jetstack/cert-manager/pkg/client/informers/externalversions",
				ListersPackage:   "github.com/jetstack/cert-manager/pkg/client/listers",
			},
			"build.knative.dev": {
				Types: []interface{}{
					v1alpha1.Build{},
				},
				PackageName:      "build",
				ClientSetPackage: "github.com/knative/build/pkg/client/clientset/versioned",
				InformersPackage: "github.com/knative/build/pkg/client/informers/externalversions",
				ListersPackage:   "github.com/knative/build/pkg/client/listers",
			},
			"networking.istio.io": {
				Types: []interface{}{
					v1alpha3.Gateway{},
					v1alpha3.VirtualService{},
					v1alpha3.DestinationRule{},
					v1alpha3.ServiceEntry{},
				},
				PackageName:      "istio",
				ClientSetPackage: "github.com/knative/pkg/client/clientset/versioned",
				InformersPackage: "github.com/knative/pkg/client/informers/externalversions",
				ListersPackage:   "github.com/knative/pkg/client/listers",
			},
			"apiextensions.k8s.io": {
				Types: []interface{}{
					v1beta1.CustomResourceDefinition{},
				},
				PackageName:      "apiextensions",
				ClientSetPackage: "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset",
				InformersPackage: "k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions",
				ListersPackage:   "k8s.io/apiextensions-apiserver/pkg/client/listers",
			},
		},
	})
}
