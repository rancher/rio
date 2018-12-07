package k8s

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/rancher/rio/pkg/deploy/stackdef/output"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	yamlDecoder "k8s.io/apimachinery/pkg/util/yaml"
)

func Populate(stack *v1.InternalStack, output *output.Deployment) error {
	if stack.Kubernetes.Manifest != "" {
		err := readManifest("", stack.Kubernetes.Manifest, output)
		if err != nil {
			return err
		}
	}

	if stack.Kubernetes.NamespacedManifest != "" {
		err := readManifest(output.Namespace, stack.Kubernetes.NamespacedManifest, output)
		if err != nil {
			return err
		}
	}

	return nil
}

func readManifest(namespace, content string, output *output.Deployment) error {
	reader := yamlDecoder.NewYAMLReader(bufio.NewReader(bytes.NewBufferString(content)))
	for {
		raw, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		objs, err := toObjects(namespace, raw)
		if err != nil {
			return err
		}

		k8sObjs, ok := output.K8sObjects[namespace]
		if !ok {
			k8sObjs = map[string]runtime.Object{}
			output.K8sObjects[namespace] = k8sObjs
		}

		for _, obj := range objs {
			k8sObjs[fmt.Sprintf("%s/%s", obj.GetName(), obj.GetKind())] = obj
		}
	}

	return nil
}

func toObjects(namespace string, raw []byte) ([]*unstructured.Unstructured, error) {
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
		return []*unstructured.Unstructured{obj}, nil
	}

	var result []*unstructured.Unstructured
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
