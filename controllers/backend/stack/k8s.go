package stack

import (
	"fmt"
	"strings"

	"bufio"
	"io"

	"bytes"

	"reflect"

	"github.com/rancher/norman/name"
	"github.com/rancher/rio/pkg/apply"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	v1beta12 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	yamlDecoder "k8s.io/apimachinery/pkg/util/yaml"
)

func deployK8sResources(stackName, namespace string, stack *v1beta1.InternalStack) error {
	var nsObjects []runtime.Object
	globalObjects := crdsForCRDDefs(true, stack.Kubernetes.NamespacedCustomResourceDefinitions)
	globalObjects = append(globalObjects, crdsForCRDDefs(false, stack.Kubernetes.CustomResourceDefinitions)...)

	if stack.Kubernetes.Manifest != "" {
		objs, err := readManifest("", stack.Kubernetes.Manifest)
		if err != nil {
			return err
		}
		globalObjects = append(globalObjects, objs...)
	}

	if stack.Kubernetes.NamespacedManifest != "" {
		objs, err := readManifest(namespace, stack.Kubernetes.Manifest)
		if err != nil {
			return err
		}
		nsObjects = append(nsObjects, objs...)
	}

	if len(globalObjects) > 0 {
		err := apply.ApplyAnyNamespace(globalObjects, fmt.Sprintf("k8s-global-%s-%s", namespace, stackName), 0)
		if err != nil {
			return err
		}
	}

	if len(nsObjects) > 0 {
		err := apply.Apply(nsObjects, fmt.Sprintf("k8s-%s", stackName), 0)
		if err != nil {
			return err
		}
	}

	return nil
}

func crdsForCRDDefs(namespaced bool, crdDefs []v1beta1.CustomResourceDefinition) []runtime.Object {
	var objs []runtime.Object
	for _, crdDef := range crdDefs {
		plural := name.GuessPluralName(strings.ToLower(crdDef.Kind))
		crdName := strings.ToLower(fmt.Sprintf("%s.%s", plural, crdDef.Group))
		crd := &v1beta12.CustomResourceDefinition{
			TypeMeta: v1.TypeMeta{
				Kind:       "CustomResourceDefinition",
				APIVersion: "apiextensions.k8s.io/v1beta1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name: crdName,
			},
			Spec: v1beta12.CustomResourceDefinitionSpec{
				Group: crdDef.Group,
				Names: v1beta12.CustomResourceDefinitionNames{
					Kind:     crdDef.Kind,
					ListKind: crdDef.Kind + "List",
					Plural:   plural,
				},
				Version: crdDef.Version,
			},
		}

		if namespaced {
			crd.Spec.Scope = v1beta12.NamespaceScoped
		}

		objs = append(objs, crd)
	}

	return objs
}

func readManifest(namespace, content string) ([]runtime.Object, error) {
	var result []runtime.Object
	reader := yamlDecoder.NewYAMLReader(bufio.NewReader(bytes.NewBufferString(content)))
	for {
		raw, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		objs, err := toObjects(namespace, raw)
		if err != nil {
			return nil, err
		}

		result = append(result, objs...)
	}

	return result, nil
}

func toObjects(namespace string, raw []byte) ([]runtime.Object, error) {
	var data map[string]interface{}
	dec := yamlDecoder.NewYAMLToJSONDecoder(bytes.NewReader(raw))
	if err := dec.Decode(&data); err != nil {
		return nil, err
	}

	listData, ok := list(data)
	if !ok {
		obj := &unstructured.Unstructured{
			Object: data,
		}
		if namespace != "" {
			obj.SetNamespace(namespace)
		}
		return []runtime.Object{obj}, nil
	}

	var result []runtime.Object
	for _, data := range listData {
		m, ok := data.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("value in manifest is expected to map[string]interface{}, got: %v", reflect.TypeOf(data))
		}

		obj := &unstructured.Unstructured{
			Object: m,
		}
		if namespace != "" {
			obj.SetNamespace(namespace)
		}

		result = append(result, obj)
	}

	return result, nil
}

func list(data map[string]interface{}) ([]interface{}, bool) {
	str, _ := data["Kind"].(string)
	items, ok := data["Items"].([]interface{})
	return items, strings.HasSuffix(str, "List") && ok
}
