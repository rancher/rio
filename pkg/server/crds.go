package server

import (
	"github.com/rancher/norman/pkg/openapi"
	rioadminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/crd"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

func getCRDs() []crd.CRD {
	crds := append([]crd.CRD{
		newCRD("ExternalService.rio.cattle.io/v1", v1.ExternalService{}),
		newCRD("Router.rio.cattle.io/v1", v1.Router{}),
		newCRD("Service.rio.cattle.io/v1", v1.Service{}),
		newCRD("Stack.rio.cattle.io/v1", v1.Stack{}),
	})

	crds = append(crds,
		newClusterCRD("ClusterDomain.admin.rio.cattle.io/v1", rioadminv1.ClusterDomain{}),
		newClusterCRD("PublicDomain.admin.rio.cattle.io/v1", rioadminv1.PublicDomain{}))

	crds = append(crds, crd.NonNamespacedTypes(
		"ClusterIssuer.certmanager.k8s.io/v1alpha1",
		"RioInfo.admin.rio.cattle.io/v1",
	)...)

	crds = append(crds, crd.NamespacedTypes(
		"GitCommit.gitwatcher.cattle.io/v1",
		"GitWatcher.gitwatcher.cattle.io/v1",

		"ServiceScaleRecommendation.autoscale.rio.cattle.io/v1",

		"Certificate.certmanager.k8s.io/v1alpha1",
		"Challenge.certmanager.k8s.io/v1alpha1",
		"Issuer.certmanager.k8s.io/v1alpha1",
		"Order.certmanager.k8s.io/v1alpha1",
	)...)

	return crds
}

func newClusterCRD(name string, obj interface{}) crd.CRD {
	return crd.NonNamespacedType(name).
		WithStatus().
		WithSchema(mustSchema(obj))
}

func newCRD(name string, obj interface{}) crd.CRD {
	return crd.NamespacedType(name).
		WithStatus().
		WithSchema(mustSchema(obj))
}

func mustSchema(obj interface{}) *v1beta1.JSONSchemaProps {
	result, err := openapi.ToOpenAPIFromStruct(obj)
	if err != nil {
		panic(err)
	}
	return result
}
