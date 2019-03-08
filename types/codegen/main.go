package main

import (
	certmanagerapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	"github.com/rancher/norman/generator"
	buildSchema "github.com/rancher/rio/types/apis/build.knative.dev/v1alpha1/schema"
	networkingSchema "github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3/schema"
	projectSchema "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1/schema"
	autoscaleSchema "github.com/rancher/rio/types/apis/rio-autoscale.cattle.io/v1/schema"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1/schema"
	webhookschema "github.com/rancher/rio/types/apis/webhookinator.rio.cattle.io/v1"
	"github.com/sirupsen/logrus"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	v1 "k8s.io/api/storage/v1"
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

var (
	basePackage = "github.com/rancher/rio/types"
)

func main() {
	if err := generator.DefaultGenerate(schema.Schemas, basePackage, true, nil); err != nil {
		logrus.Fatal(err)
	}
	if err := generator.DefaultGenerate(networkingSchema.Schemas, basePackage, false, nil); err != nil {
		logrus.Fatal(err)
	}
	if err := generator.DefaultGenerate(projectSchema.Schemas, basePackage, true, nil); err != nil {
		logrus.Fatal(err)
	}
	if err := generator.DefaultGenerate(autoscaleSchema.Schemas, basePackage, false, nil); err != nil {
		logrus.Fatal(err)
	}
	if err := generator.DefaultGenerate(buildSchema.Schemas, basePackage, false, nil); err != nil {
		logrus.Fatal(err)
	}
	if err := generator.DefaultGenerate(webhookschema.Schemas, basePackage, false, nil); err != nil {
		logrus.Fatal(err)
	}
	if err := generator.ControllersForForeignTypes(basePackage, extv1beta1.SchemeGroupVersion,
		nil,
		[]interface{}{
			extv1beta1.CustomResourceDefinition{},
		}); err != nil {
		logrus.Fatal(err)
	}
	if err := generator.ControllersForForeignTypes(basePackage, v1.SchemeGroupVersion,
		nil,
		[]interface{}{
			v1.StorageClass{},
		}); err != nil {
		logrus.Fatal(err)
	}
	if err := generator.ControllersForForeignTypes(basePackage, policyv1beta1.SchemeGroupVersion,
		[]interface{}{
			policyv1beta1.PodDisruptionBudget{},
		}, nil); err != nil {
		logrus.Fatal(err)
	}
	if err := generator.ControllersForForeignTypes(basePackage, certmanagerapi.SchemeGroupVersion,
		[]interface{}{
			certmanagerapi.Certificate{},
		},
		[]interface{}{
			certmanagerapi.ClusterIssuer{},
		},
	); err != nil {
		logrus.Fatal(err)
	}
}
