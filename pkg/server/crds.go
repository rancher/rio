package server

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/rancher/wrangler/pkg/kv"

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
		"RioInfo.admin.rio.cattle.io/v1",
	)...)

	crds = append(crds, crd.NamespacedTypes(
		"GitCommit.gitwatcher.cattle.io/v1",
		"GitWatcher.gitwatcher.cattle.io/v1",
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
		WithSchema(mustSchema(obj)).
		WithCustomColumn(customColumn(obj))
}

type customResourceColumnDefinitionList struct {
	list []v1beta1.CustomResourceColumnDefinition
}

func customColumn(obj interface{}) []v1beta1.CustomResourceColumnDefinition {
	var r customResourceColumnDefinitionList
	t := reflect.TypeOf(obj)
	readCustomColumn(t, &r)
	return append(r.list, v1beta1.CustomResourceColumnDefinition{
		Name:     "Age",
		Type:     "date",
		JSONPath: ".metadata.creationTimestamp",
	})
}

func readCustomColumn(t reflect.Type, r *customResourceColumnDefinitionList) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i).Type
		if f.Kind() == reflect.Ptr {
			f = f.Elem()
		}
		if f.Kind() == reflect.Struct {
			readCustomColumn(f, r)
		} else {
			c := v1beta1.CustomResourceColumnDefinition{}
			kvs := strings.Split(t.Field(i).Tag.Get("column"), ",")
			for _, keyValues := range kvs {
				k, v := kv.Split(keyValues, "=")
				switch k {
				case "name":
					c.Name = v
				case "type":
					c.Type = v
				case "format":
					c.Format = v
				case "description":
					c.Description = v
				case "priority":
					p, _ := strconv.Atoi(v)
					c.Priority = int32(p)
				case "jsonpath":
					c.JSONPath = v
				}
			}
			if c.Name == "" {
				continue
			}
			r.list = append(r.list, c)
		}
	}
}

func mustSchema(obj interface{}) *v1beta1.JSONSchemaProps {
	result, err := openapi.ToOpenAPIFromStruct(obj)
	if err != nil {
		panic(err)
	}
	return result
}
