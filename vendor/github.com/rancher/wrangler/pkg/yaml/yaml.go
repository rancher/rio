package yaml

import (
	"bufio"
	"bytes"
	"io"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	yamlDecoder "k8s.io/apimachinery/pkg/util/yaml"
)

func Unmarshal(data []byte, v interface{}) error {
	return yamlDecoder.NewYAMLToJSONDecoder(bytes.NewBuffer(data)).Decode(v)
}

func ToObjects(in io.Reader) ([]runtime.Object, error) {
	var result []runtime.Object
	reader := yamlDecoder.NewYAMLReader(bufio.NewReaderSize(in, 4096))
	for {
		raw, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		obj, err := toObjects(raw)
		if err != nil {
			return nil, err
		}

		result = append(result, obj...)
	}

	return result, nil
}

func toObjects(bytes []byte) ([]runtime.Object, error) {
	bytes, err := yamlDecoder.ToJSON(bytes)
	if err != nil {
		return nil, err
	}
	obj, _, err := unstructured.UnstructuredJSONScheme.Decode(bytes, nil, nil)
	if err != nil {
		return nil, err
	}

	if l, ok := obj.(*unstructured.UnstructuredList); ok {
		var result []runtime.Object
		for _, obj := range l.Items {
			copy := obj
			result = append(result, &copy)
		}
		return result, nil
	}

	return []runtime.Object{obj}, nil
}

func ToBytes(objects []runtime.Object) ([]byte, error) {
	if len(objects) == 0 {
		return nil, nil
	}

	buffer := &bytes.Buffer{}
	for i, obj := range objects {
		if i > 0 {
			buffer.WriteString("\n---\n")
		}

		bytes, err := yaml.Marshal(obj)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to encode %s", obj.GetObjectKind().GroupVersionKind())
		}
		buffer.Write(bytes)
	}

	return buffer.Bytes(), nil
}
